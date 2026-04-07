package register

import (
	"context"
	"encoding/json"
	"expire-share/internal/domain/dto/auth/commands"
	"expire-share/internal/domain/dto/auth/results"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/mocks"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Register(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	validReq := Request{
		Login:    "mkaascs",
		Email:    "mkaascs@gmail.com",
		Password: "password123",
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRegister := mocks.NewMockUserRegister(ctrl)
		mockRegister.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cmd commands.Register) (*results.Register, error) {
				require.Equal(t, validReq.Login, cmd.Login)
				require.Equal(t, validReq.Email, cmd.Email)
				require.Equal(t, validReq.Password, cmd.Password)
				return &results.Register{UserID: 1}, nil
			})

		handler := New(mockRegister, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRegisterRequest(validReq))

		require.Equal(t, http.StatusOK, w.Code)

		var resp Response
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		require.Equal(t, int64(1), resp.UserID)
	})

	t.Run("missing request in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRegister := mocks.NewMockUserRegister(ctrl)
		handler := New(mockRegister, logger)

		r := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRegister := mocks.NewMockUserRegister(ctrl)
		mockRegister.EXPECT().Register(gomock.Any(), gomock.Any()).
			Return(nil, domainErrors.ErrUserAlreadyExists)

		handler := New(mockRegister, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRegisterRequest(validReq))

		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRegister := mocks.NewMockUserRegister(ctrl)
		mockRegister.EXPECT().Register(gomock.Any(), gomock.Any()).
			Return(nil, context.Canceled)

		handler := New(mockRegister, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRegisterRequest(validReq))

		require.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRegister := mocks.NewMockUserRegister(ctrl)
		mockRegister.EXPECT().Register(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("db error"))

		handler := New(mockRegister, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRegisterRequest(validReq))

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func newRegisterRequest(req Request) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	ctx := context.WithValue(r.Context(), "request", req)
	return r.WithContext(ctx)
}
