package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/witelokk/music-api/internal/requestctx"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

var ErrUnauthorized = errors.New("unauthorized")

func NewJWTMiddleware(jwtSecret string, logger *slog.Logger) openapi.StrictMiddlewareFunc {
	return func(next openapi.StrictHandlerFunc, operationID string) openapi.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			reqLogger := requestctx.LoggerFromContext(ctx, logger)

			requiresAuth := func(op string) bool {
				switch op {
				case "GetCurrentUser", "GetSong", "GetArtist", "GetRelease":
					return true
				default:
					return false
				}
			}

			if !requiresAuth(operationID) {
				return next(ctx, w, r, request)
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				if operationID == "GetCurrentUser" {
					return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "missing or invalid authorization header"}), nil
				}
				return nil, ErrUnauthorized
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "" {
				if operationID == "GetCurrentUser" {
					return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "missing or invalid authorization header"}), nil
				}
				return nil, ErrUnauthorized
			}

			parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
				if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !parsed.Valid {
				if err != nil {
					reqLogger.Warn("invalid access token",
						slog.String("error", err.Error()),
					)
				} else {
					reqLogger.Warn("invalid access token")
				}
				if operationID == "GetCurrentUser" {
					return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "invalid or expired access token"}), nil
				}
				return nil, ErrUnauthorized
			}

			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				reqLogger.Warn("invalid access token claims")
				if operationID == "GetCurrentUser" {
					return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "invalid or expired access token"}), nil
				}
				return nil, ErrUnauthorized
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				reqLogger.Warn("missing sub claim in access token")
				if operationID == "GetCurrentUser" {
					return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "invalid or expired access token"}), nil
				}
				return nil, ErrUnauthorized
			}

			ctxWithUser := WithUserID(ctx, sub)
			return next(ctxWithUser, w, r, request)
		}
	}
}
