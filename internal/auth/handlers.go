package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/witelokk/music-api/internal/requestctx"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func HandleSendVerificationEmail(
	ctx context.Context,
	service *AuthService,
	logger *slog.Logger,
	req openapi.SendVerificationEmailRequestObject,
) (openapi.SendVerificationEmailResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)
	if req.Body == nil {
		return openapi.SendVerificationEmail400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	body := req.Body

	if string(body.Email) == "" {
		return openapi.SendVerificationEmail400JSONResponse(openapi.Error{Error: "email is required"}), nil
	}

	if _, err := mail.ParseAddress(string(body.Email)); err != nil {
		return openapi.SendVerificationEmail400JSONResponse(openapi.Error{Error: "invalid email"}), nil
	}

	if err := service.SendVerificationEmail(ctx, string(body.Email)); err != nil {
		if errors.Is(err, ErrVerificationCodeRecentlySent) {
			return openapi.SendVerificationEmail429JSONResponse(openapi.Error{Error: "verification code recently sent"}), nil
		}

		reqLogger.Error("failed to send verification email",
			slog.String("email", string(body.Email)),
			slog.String("error", err.Error()),
		)

		return openapi.SendVerificationEmail500JSONResponse(openapi.Error{Error: "failed to send verification email"}), nil
	}

	return openapi.SendVerificationEmail204Response{}, nil
}

func HandleCreateUser(
	ctx context.Context,
	service *AuthService,
	logger *slog.Logger,
	req openapi.CreateUserRequestObject,
) (openapi.CreateUserResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)
	if req.Body == nil {
		return openapi.CreateUser400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	body := req.Body

	if body.Name == "" {
		return openapi.CreateUser400JSONResponse(openapi.Error{Error: "name is required"}), nil
	}

	if string(body.Email) == "" {
		return openapi.CreateUser400JSONResponse(openapi.Error{Error: "email is required"}), nil
	}

	if _, err := mail.ParseAddress(string(body.Email)); err != nil {
		return openapi.CreateUser400JSONResponse(openapi.Error{Error: "invalid email"}), nil
	}

	if body.Code == "" {
		return openapi.CreateUser400JSONResponse(openapi.Error{Error: "verification_code is required"}), nil
	}

	user := &User{
		Name:  body.Name,
		Email: string(body.Email),
	}

	if err := service.CreateUser(ctx, user, body.Code); err != nil {
		switch {
		case errors.Is(err, ErrInvalidVerificationCode):
			return openapi.CreateUser400JSONResponse(openapi.Error{Error: "invalid verification code"}), nil
		case errors.Is(err, ErrExpiredVerificationCode):
			return openapi.CreateUser400JSONResponse(openapi.Error{Error: "verification code expired"}), nil
		case errors.Is(err, ErrUserAlreadyExists):
			return openapi.CreateUser409JSONResponse(openapi.Error{Error: "user already exists"}), nil
		default:
			reqLogger.Error("failed to create user",
				slog.String("email", string(body.Email)),
				slog.String("error", err.Error()),
			)
			return openapi.CreateUser500JSONResponse(openapi.Error{Error: "failed to create user"}), nil
		}
	}

	resp := openapi.User{
		Id:        uuid.MustParse(user.ID),
		Name:      user.Name,
		Email:     openapi_types.Email(user.Email),
		CreatedAt: user.CreatedAt,
	}

	return openapi.CreateUser201JSONResponse(resp), nil
}

func HandleGenerateTokens(
	ctx context.Context,
	service *AuthService,
	logger *slog.Logger,
	req openapi.GenerateTokensRequestObject,
) (openapi.GenerateTokensResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)
	if req.Body == nil {
		return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	body := req.Body

	switch body.GrantType {
	case openapi.Code:
		if body.Email == nil || *body.Email == "" {
			return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "email is required"}), nil
		}
		if body.Code == nil || *body.Code == "" {
			return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "verification code is required"}), nil
		}

		tokens, err := service.GetTokensWithVerificationCode(ctx, string(*body.Email), *body.Code)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidVerificationCode):
				return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "invalid verification code"}), nil
			case errors.Is(err, ErrExpiredVerificationCode):
				return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "verification code expired"}), nil
			default:
				reqLogger.Error("failed to get tokens with verification code",
					slog.String("email", string(*body.Email)),
					slog.String("error", err.Error()),
				)
				return openapi.GenerateTokens500JSONResponse(openapi.Error{Error: "failed to get tokens"}), nil
			}
		}

		return openapi.GenerateTokens200JSONResponse(openapi.TokensResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		}), nil

	case openapi.RefreshToken:
		if body.RefreshToken == nil || *body.RefreshToken == "" {
			return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "refresh_token is required"}), nil
		}

		tokens, err := service.GetTokensWithRefreshToken(ctx, *body.RefreshToken)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidRefreshToken):
				return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "invalid refresh token"}), nil
			case errors.Is(err, ErrExpiredRefreshToken):
				return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "refresh token expired"}), nil
			default:
				reqLogger.Error("failed to get tokens with refresh token",
					slog.String("error", err.Error()),
				)
				return openapi.GenerateTokens500JSONResponse(openapi.Error{Error: "failed to get tokens"}), nil
			}
		}

		return openapi.GenerateTokens200JSONResponse(openapi.TokensResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		}), nil

	default:
		return openapi.GenerateTokens400JSONResponse(openapi.Error{Error: "unsupported grant_type"}), nil
	}
}

func HandleGetCurrentUser(
	ctx context.Context,
	service *AuthService,
	logger *slog.Logger,
	req openapi.GetCurrentUserRequestObject,
) (openapi.GetCurrentUserResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	user, err := service.GetCurrentUser(ctx)
	if err != nil {
		if errors.Is(err, ErrInvalidAccessToken) {
			reqLogger.Warn("invalid or missing user in context for GetCurrentUser",
				slog.String("error", err.Error()),
			)
			return openapi.GetCurrentUser401JSONResponse(openapi.Error{Error: "unauthorized"}), nil
		}

		reqLogger.Error("failed to get current user",
			slog.String("error", err.Error()),
		)
		return openapi.GetCurrentUser500JSONResponse(openapi.Error{Error: "failed to fetch current user"}), nil
	}

	resp := openapi.CurrentUser{
		Name:  user.Name,
		Email: openapi_types.Email(user.Email),
	}

	return openapi.GetCurrentUser200JSONResponse(resp), nil
}
