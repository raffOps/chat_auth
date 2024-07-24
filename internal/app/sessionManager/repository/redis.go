package sessionManager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/raffops/auth/internal/app/sessionManager"
	"github.com/raffops/chat/pkg/encryptor"
	"github.com/raffops/chat/pkg/errs"
	"github.com/redis/go-redis/v9"
	"time"
)

type redisRepository struct {
	db        *redis.Client
	encryptor encryptor.Encryptor
}

func (r redisRepository) GetEncryptedHashmap(ctx context.Context, key, secretKey string) (
	map[string]interface{},
	errs.ChatError,
) {
	encryptedValue, err := r.db.Get(ctx, key).Result()
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, err)
	}

	decryptedValue, err := r.encryptor.Decrypt(encryptedValue, secretKey)
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, err)
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(decryptedValue), &result)
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, fmt.Errorf("corrupted data"))
	}

	return result, nil
}

func (r redisRepository) SetEncryptedHashmap(
	ctx context.Context,
	key, secretKey string,
	value map[string]interface{},
) errs.ChatError {
	valueByte, err := json.Marshal(value)
	if err != nil {
		return errs.NewError(
			errs.ErrInternal,
			fmt.Errorf("error marshaling value: %v", err),
		)
	}
	encryptedValue, err := r.encryptor.Encrypt(string(valueByte), secretKey)
	if err != nil {
		return errs.NewError(
			errs.ErrInternal,
			fmt.Errorf("error encrypting value: %v", err),
		)
	}
	err = r.db.Set(ctx, key, encryptedValue, 0).Err()
	if err != nil {
		return errs.NewError(
			errs.ErrInternal,
			fmt.Errorf("error creating session: %v", err),
		)
	}
	return nil
}

func (r redisRepository) ExpireAt(ctx context.Context, key string, at time.Time) errs.ChatError {
	err := r.db.ExpireAt(ctx, key, at).Err()
	if err != nil {
		return errs.NewError(
			errs.ErrInternal,
			fmt.Errorf("error setting ttl: %v", err),
		)
	}
	return nil
}

func (r redisRepository) Hashset(ctx context.Context, key string, value map[string]interface{}) errs.ChatError {
	for k, v := range value {
		_, err := r.db.HSet(ctx, key, k, v).Result()
		if err != nil {
			return errs.NewError(
				errs.ErrInternal,
				fmt.Errorf("error creating session: %v", err),
			)
		}
	}
	return nil
}

func (r redisRepository) SetAppend(ctx context.Context, key, value string) errs.ChatError {
	err := r.db.SAdd(ctx, key, value).Err()
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return nil
}

func (r redisRepository) Hashget(ctx context.Context, key, field string) (string, errs.ChatError) {
	result, err := r.db.HGet(ctx, key, field).Result()
	if err != nil {
		return "", errs.NewError(errs.ErrInternal, err)
	}
	return result, nil
}

func NewRedisRepository(db *redis.Client, encryptor encryptor.Encryptor) sessionManager.Repository {
	return &redisRepository{db: db, encryptor: encryptor}
}
