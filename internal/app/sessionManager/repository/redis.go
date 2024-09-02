package sessionManager

import (
	"context"
	"encoding/json"
	"errors"
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

func (r redisRepository) GetKeys(ctx context.Context, pattern string) ([]string, errs.ChatError) {
	keys, err := r.db.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, err)
	}
	return keys, nil
}

func (r redisRepository) BeginTransaction(ctx context.Context) (interface{}, errs.ChatError) {
	defer func() (interface{}, errs.ChatError) {
		if r := recover(); r != nil {
			return nil, errs.NewError(
				errs.ErrInternal,
				errors.New("error starting transaction"),
			)
		}
		return nil, nil
	}()
	return r.db.TxPipeline(), nil
}

func (r redisRepository) CommitTransaction(ctx context.Context, tx interface{}) errs.ChatError {
	db := tx.(redis.Pipeliner)
	_, err := db.Exec(ctx)
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return nil
}

func (r redisRepository) RollbackTransaction(ctx context.Context, tx interface{}) errs.ChatError {
	tx.(redis.Pipeliner).Discard()
	return nil
}

func (r redisRepository) StringGet(ctx context.Context, tableName, key string) (string, errs.ChatError) {
	id := fmt.Sprintf("%s:%s", tableName, key)
	value, err := r.db.Get(ctx, id).Result()
	if err != nil {
		// TODO improve error handling
		return "", errs.NewError(errs.ErrInternal, err)
	}
	return value, nil
}

func (r redisRepository) StringSet(ctx context.Context, tx interface{}, tableName, key, value string) errs.ChatError {
	id := fmt.Sprintf("%s:%s", tableName, key)
	db := tx.(redis.Pipeliner)
	err := db.Set(ctx, id, value, 0).Err()
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return nil
}

func (r redisRepository) Delete(ctx context.Context, tx interface{}, tableName, key string) errs.ChatError {
	id := fmt.Sprintf("%s:%s", tableName, key)
	db := tx.(redis.Pipeliner)
	err := db.Del(ctx, id).Err()
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return nil
}

func (r redisRepository) HashGet(
	ctx context.Context,
	tableName, key string,
	columns ...string,
) (map[string]interface{}, errs.ChatError) {
	id := fmt.Sprintf("%s:%s", tableName, key)
	output := make(map[string]interface{})
	var err error
	for _, column := range columns {
		output[column], err = r.db.HGet(ctx, id, column).Result()
		if err != nil {
			return nil, errs.NewError(errs.ErrInternal, err)
		}
	}
	return output, nil
}

func (r redisRepository) HashGetEncrypted(
	ctx context.Context,
	tableName, key, secret string,
) (map[string]interface{}, errs.ChatError) {
	encryptedValues, err := r.HashGet(ctx, tableName, key, "encrypted_value")
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, err)
	}
	encryptedValue := encryptedValues["encrypted_value"].(string)
	decryptedValue, errDecrypt := r.encryptor.Decrypt(encryptedValue, secret)
	if errDecrypt != nil {
		return nil, errs.NewError(errs.ErrInternal, errDecrypt)
	}

	var output map[string]interface{}
	errUnmarshal := json.Unmarshal([]byte(decryptedValue), &output)
	if errUnmarshal != nil {
		return nil, errs.NewError(errs.ErrInternal, errUnmarshal)
	}

	return output, nil
}

func (r redisRepository) HashSet(
	ctx context.Context,
	tx interface{},
	tableName, key string,
	values map[string]interface{},
) errs.ChatError {
	id := fmt.Sprintf("%s:%s", tableName, key)
	db := tx.(redis.Pipeliner)
	err := db.HSet(ctx, id, values).Err()
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return nil
}

func (r redisRepository) HashSetEncrypted(
	ctx context.Context,
	tx interface{},
	tableName, key, secret string,
	values map[string]interface{},
) errs.ChatError {
	valueByte, err := json.Marshal(values)
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	encryptedValue, err := r.encryptor.Encrypt(string(valueByte), secret)
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return r.HashSet(ctx, tx, tableName, key, map[string]interface{}{"encrypted_value": encryptedValue})
}

func (r redisRepository) GetTTL(ctx context.Context, tableName, key string) (time.Time, errs.ChatError) {
	id := fmt.Sprintf("%s:%s", tableName, key)
	ttl := r.db.ExpireTime(ctx, id).Val().Seconds()
	return time.Unix(int64(ttl), 0), nil
}

func (r redisRepository) ExpireAt(
	ctx context.Context,
	tx interface{},
	tableName string,
	key string,
	at time.Time,
) errs.ChatError {
	id := fmt.Sprintf("%s:%s", tableName, key)
	db := tx.(redis.Pipeliner)
	err := db.ExpireAt(ctx, id, at).Err()
	if err != nil {
		return errs.NewError(errs.ErrInternal, err)
	}
	return nil
}

func NewRedisRepository(db *redis.Client, encryptor encryptor.Encryptor) sessionManager.ReaderWriterRepository {
	return &redisRepository{db: db, encryptor: encryptor}
}
