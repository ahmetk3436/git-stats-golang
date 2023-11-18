package storage

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(address, password string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // Dockerfile-Redis veritabanı seçimi
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisClient{client}, nil
}

func (rc *RedisClient) Set(key string, value interface{}, duration time.Duration) error {
	err := rc.client.Set(key, value, duration*time.Minute).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisClient) Get(key string) ([]byte, error) {
	val, err := rc.client.Get(key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, errors.New(err.Error())
	}
	return val, nil
}

func (rc *RedisClient) Delete(key string) error {
	err := rc.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}
