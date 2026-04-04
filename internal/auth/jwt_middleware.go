package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func NewJWTMiddleware(jwtSecret string, logger *slog.Logger) openapi.StrictMiddlewareFunc {
	return func(next openapi.StrictHandlerFunc, operationID string) openapi.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			if operationID != "GetCurrentUser" {
				return next(ctx, w, r, request)
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "missing or invalid authorization header"}), nil
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "" {
				return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "missing or invalid authorization header"}), nil
			}

			parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
				if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !parsed.Valid {
				logger.Warn("invalid access token",
					slog.String("error", err.Error()),
				)
				return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "invalid or expired access token"}), nil
			}

			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("invalid access token claims")
				return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "invalid or expired access token"}), nil
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				logger.Warn("missing sub claim in access token")
				return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "invalid or expired access token"}), nil
			}

			ctxWithUser := WithUserID(ctx, sub)
			return next(ctxWithUser, w, r, request)
		}
	}
}
