package refresh

import (
	"context"
	"encoding/json"
	"expire-share/internal/domain/dto/auth/commands"
	"expire-share/internal/domain/dto/auth/results"
	"expire-share/internal/domain/entities"
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

func TestHandler_Refresh(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	validReq := Request{RefreshToken: "refresh-token"}
	validResult := &results.Refresh{
		Tokens: entities.TokenPair{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
		},
		ExpiresIn: 900,
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRefresh := mocks.NewMockTokenRefresh(ctrl)
		mockRefresh.EXPECT().
			Refresh(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cmd commands.Refresh) (*results.Refresh, error) {
				require.Equal(t, validReq.RefreshToken, cmd.RefreshToken)
				return validResult, nil
			})

		handler := New(mockRefresh, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRefreshRequest(validReq))

		require.Equal(t, http.StatusOK, w.Code)

		var resp Response
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		require.Equal(t, "new-access-token", resp.AccessToken)
		require.Equal(t, "new-refresh-token", resp.RefreshToken)
		require.Equal(t, int64(900), resp.ExpiresIn)
		require.NotEqual(t, validReq.RefreshToken, resp.RefreshToken)
	})

	t.Run("missing request in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRefresh := mocks.NewMockTokenRefresh(ctrl)
		handler := New(mockRefresh, logger)

		r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRefresh := mocks.NewMockTokenRefresh(ctrl)
		mockRefresh.EXPECT().Refresh(gomock.Any(), gomock.Any()).
			Return(nil, domainErrors.ErrInvalidRefreshToken)

		handler := New(mockRefresh, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRefreshRequest(validReq))

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRefresh := mocks.NewMockTokenRefresh(ctrl)
		mockRefresh.EXPECT().Refresh(gomock.Any(), gomock.Any()).
			Return(nil, context.Canceled)

		handler := New(mockRefresh, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRefreshRequest(validReq))

		require.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRefresh := mocks.NewMockTokenRefresh(ctrl)
		mockRefresh.EXPECT().Refresh(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("db error"))

		handler := New(mockRefresh, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRefreshRequest(validReq))

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func newRefreshRequest(req Request) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	ctx := context.WithValue(r.Context(), "request", req)
	return r.WithContext(ctx)
}
