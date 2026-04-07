package login

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

func TestHandler_Login(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	validReq := Request{Login: "mkaascs", Password: "password123"}
	validResult := &results.Login{
		Tokens: entities.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},

		ExpiresIn: 900,
		User:      entities.User{ID: 1, Login: "mkaascs"},
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogin := mocks.NewMockUserLogin(ctrl)
		mockLogin.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cmd commands.Login) (*results.Login, error) {
				require.Equal(t, validReq.Login, cmd.Login)
				require.Equal(t, validReq.Password, cmd.Password)
				return validResult, nil
			})

		handler := New(mockLogin, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLoginRequest(validReq))

		require.Equal(t, http.StatusOK, w.Code)

		var resp Response
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		require.Equal(t, "access-token", resp.AccessToken)
		require.Equal(t, "refresh-token", resp.RefreshToken)
		require.Equal(t, int64(900), resp.ExpiresIn)
	})

	t.Run("missing request in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogin := mocks.NewMockUserLogin(ctrl)
		handler := New(mockLogin, logger)

		r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogin := mocks.NewMockUserLogin(ctrl)
		mockLogin.EXPECT().Login(gomock.Any(), gomock.Any()).
			Return(nil, domainErrors.ErrInvalidCredentials)

		handler := New(mockLogin, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLoginRequest(validReq))

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogin := mocks.NewMockUserLogin(ctrl)
		mockLogin.EXPECT().Login(gomock.Any(), gomock.Any()).
			Return(nil, context.Canceled)

		handler := New(mockLogin, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLoginRequest(validReq))

		require.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogin := mocks.NewMockUserLogin(ctrl)
		mockLogin.EXPECT().Login(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("db error"))

		handler := New(mockLogin, logger)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newLoginRequest(validReq))

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func newLoginRequest(req Request) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	ctx := context.WithValue(r.Context(), "request", req)
	return r.WithContext(ctx)
}
