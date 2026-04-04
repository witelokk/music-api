package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenRepository interface {
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type RedisRefreshTokenRepository struct {
	client redis.Cmdable
}

func NewRedisRefreshTokenRepository(client redis.Cmdable) *RedisRefreshTokenRepository {
	return &RedisRefreshTokenRepository{client: client}
}

func (r *RedisRefreshTokenRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	key := fmt.Sprintf("auth:refresh_token:%s", token.Token)

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRefreshTokenRepository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	key := fmt.Sprintf("auth:refresh_token:%s", token)

	raw, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var rt RefreshToken
	if err := json.Unmarshal(raw, &rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (r *RedisRefreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("auth:refresh_token:%s", token)
	return r.client.Del(ctx, key).Err()
}
