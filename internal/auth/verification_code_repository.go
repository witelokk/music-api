package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type VerificationCodeRepository interface {
	SaveCode(ctx context.Context, code *VerificationCode) error
	GetCodesByEmail(ctx context.Context, email string) ([]*VerificationCode, error)
	DeleteCode(ctx context.Context, email string, code string) error
}

type RedisVerificationCodeRepository struct {
	client redis.Cmdable
}

func NewRedisVerificationCodeRepository(client redis.Cmdable) *RedisVerificationCodeRepository {
	return &RedisVerificationCodeRepository{client: client}
}

func (r *RedisVerificationCodeRepository) SaveCode(ctx context.Context, code *VerificationCode) error {
	key := fmt.Sprintf("auth:verification_code:%s:%s", code.Email, code.Code)
	setKey := fmt.Sprintf("auth:verification_codes:%s", code.Email)

	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	ttl := time.Until(code.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, setKey, code.Code)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisVerificationCodeRepository) GetCodesByEmail(ctx context.Context, email string) ([]*VerificationCode, error) {
	setKey := fmt.Sprintf("auth:verification_codes:%s", email)

	codes, err := r.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, err
	}

	var result []*VerificationCode

	for _, c := range codes {
		key := fmt.Sprintf("auth:verification_code:%s:%s", email, c)
		raw, err := r.client.Get(ctx, key).Bytes()
		if err == redis.Nil {
			// cleanup dangling entry
			_, _ = r.client.SRem(ctx, setKey, c).Result()
			continue
		}
		if err != nil {
			return nil, err
		}

		var vc VerificationCode
		if err := json.Unmarshal(raw, &vc); err != nil {
			return nil, err
		}
		result = append(result, &vc)
	}

	return result, nil
}

func (r *RedisVerificationCodeRepository) DeleteCode(ctx context.Context, email string, code string) error {
	key := fmt.Sprintf("auth:verification_code:%s:%s", email, code)
	setKey := fmt.Sprintf("auth:verification_codes:%s", email)

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, setKey, code)
	_, err := pipe.Exec(ctx)
	return err
}
