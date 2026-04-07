package upload_test

import (
	"bytes"
	"context"
	"encoding/json"
	"expire-share/internal/config"
	"expire-share/internal/delivery/handlers/api/upload"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/domain/entities"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/mocks"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_Upload(t *testing.T) {
	testCfg := config.Config{
		Storage: config.Storage{
			MaxFileSizeInBytes: 10 * 1024 * 1024,
		},

		Service: config.Service{
			MaxDownloads: 5,
			DefaultTtl:   2 * time.Hour,
		},
	}

	claims := &middlewares.UserClaims{
		UserID: 1,
		Roles:  []entities.UserRole{entities.RoleUser},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)

		mockUploader.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cmd commands.UploadFile) (string, error) {
				require.Equal(t, "test.txt", cmd.Filename)
				require.Equal(t, int64(1), cmd.RequestingUserInfo.UserID)
				require.Equal(t, "2h0m0s", cmd.TTL.String())
				return "abc123", nil
			})

		req := upload.Request{TTL: "2h", MaxDownloads: 5}
		r := buildMultipartRequest(t, "test.txt", "hello world")
		r = withContext(r, req, claims)

		handler := upload.New(mockUploader, logger, testCfg)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		resp := parseResponse(t, w)
		require.Equal(t, "abc123", resp.Alias)
	})

	t.Run("success with password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)

		mockUploader.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cmd commands.UploadFile) (string, error) {
				require.Equal(t, "secret", cmd.Password)
				return "xyz789", nil
			})

		req := upload.Request{TTL: "1h", MaxDownloads: 3, Password: "secret"}
		r := buildMultipartRequest(t, "doc.pdf", "pdf content")
		r = withContext(r, req, claims)

		handler := upload.New(mockUploader, logger, testCfg)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusCreated, w.Code)
		resp := parseResponse(t, w)
		require.Equal(t, "xyz789", resp.Alias)
	})

	t.Run("missing parsed request in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		handler := upload.New(mockUploader, logger, testCfg)

		r := buildMultipartRequest(t, "test.txt", "data")
		ctx := context.WithValue(r.Context(), "user_id", int64(1))
		ctx = context.WithValue(ctx, "roles", []entities.UserRole{entities.RoleUser})
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("missing user claims in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		handler := upload.New(mockUploader, logger, testCfg)

		req := upload.Request{TTL: "1h", MaxDownloads: 1}
		r := buildMultipartRequest(t, "test.txt", "data")

		ctx := context.WithValue(r.Context(), "request", req)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid ttl format", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		handler := upload.New(mockUploader, logger, testCfg)

		req := upload.Request{TTL: "not-a-duration", MaxDownloads: 1}
		r := buildMultipartRequest(t, "test.txt", "data")
		r = withContext(r, req, claims)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing file in form", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		handler := upload.New(mockUploader, logger, testCfg)

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		require.NoError(t, mw.Close())

		r := httptest.NewRequest(http.MethodPost, "/upload", &buf)
		r.Header.Set("Content-Type", mw.FormDataContentType())

		req := upload.Request{TTL: "1h", MaxDownloads: 1}
		r = withContext(r, req, claims)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("file size too big", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		mockUploader.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()).
			Return("", domainErrors.ErrFileSizeTooBig)

		req := upload.Request{TTL: "1h", MaxDownloads: 1}
		r := buildMultipartRequest(t, "big.bin", "lots of data")
		r = withContext(r, req, claims)

		handler := upload.New(mockUploader, logger, testCfg)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		mockUploader.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()).
			Return("", context.Canceled)

		req := upload.Request{TTL: "1h", MaxDownloads: 1}
		r := buildMultipartRequest(t, "test.txt", "data")

		ctx, cancel := context.WithCancel(r.Context())
		cancel()

		r = r.WithContext(ctx)
		r = withContext(r, req, claims)

		handler := upload.New(mockUploader, logger, testCfg)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUploader := mocks.NewMockFileUploader(ctrl)
		mockUploader.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("internal server error"))

		req := upload.Request{TTL: "1h", MaxDownloads: 1}
		r := buildMultipartRequest(t, "test.txt", "data")
		r = withContext(r, req, claims)

		handler := upload.New(mockUploader, logger, testCfg)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func withContext(r *http.Request, req upload.Request, claims *middlewares.UserClaims) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, "request", req)
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "roles", claims.Roles)
	return r.WithContext(ctx)
}

func buildMultipartRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = fw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())

	routeCtx := chi.NewRouteContext()
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) upload.Response {
	t.Helper()
	var resp upload.Response
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return resp
}
