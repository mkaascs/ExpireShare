package logout

import (
	"context"
	"expire-share/internal/domain/dto/auth/commands"
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

func TestHandler_Logout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	validReq := Request{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogout := mocks.NewMockUserLogout(ctrl)
		mockLogout.EXPECT().
			Logout(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cmd commands.Logout) error {
				require.Equal(t, validReq.AccessToken, cmd.AccessToken)
				require.Equal(t, validReq.RefreshToken, cmd.RefreshToken)
				return nil
			})

		handler := New(mockLogout, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLogoutRequest(validReq))

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing request in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogout := mocks.NewMockUserLogout(ctrl)
		handler := New(mockLogout, logger)

		r := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogout := mocks.NewMockUserLogout(ctrl)
		mockLogout.EXPECT().Logout(gomock.Any(), gomock.Any()).
			Return(domainErrors.ErrInvalidRefreshToken)

		handler := New(mockLogout, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLogoutRequest(validReq))

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogout := mocks.NewMockUserLogout(ctrl)
		mockLogout.EXPECT().Logout(gomock.Any(), gomock.Any()).
			Return(context.Canceled)

		handler := New(mockLogout, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLogoutRequest(validReq))

		require.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogout := mocks.NewMockUserLogout(ctrl)
		mockLogout.EXPECT().Logout(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("redis error"))

		handler := New(mockLogout, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLogoutRequest(validReq))

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func newLogoutRequest(req Request) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	ctx := context.WithValue(r.Context(), "request", req)
	return r.WithContext(ctx)
}
