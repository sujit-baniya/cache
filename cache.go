package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/storage/memory"
)

var ctx = context.Background()

type Config struct {
	Driver   string `yaml:"driver" env:"CACHE_DRIVER"`
	Name     string `yaml:"name" env:"CACHE_NAME"`
	Host     string `yaml:"host" env:"CACHE_HOST"`
	Password string `yaml:"password" env:"CACHE_PASSWORD"`
	Port     int    `yaml:"port" env:"CACHE_PORT"`
	DB       int    `yaml:"db" env:"CACHE_DB"`
}

type Cache struct {
	Memory *memory.Storage
	Redis  *redis.Client
}

var DefaultCache = &Cache{
	Memory: memory.New(),
	Redis: redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}),
}

func Default(cfg Config) {
	DefaultCache = New(cfg)
}

func New(cfg ...Config) *Cache {
	cs := &Cache{Memory: memory.New()}
	if len(cfg) == 0 {
		return DefaultCache
	}
	c := cfg[0]
	cs.Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
	})
	return cs
}

func Set(key string, value []byte, ttl time.Duration) error {
	status := DefaultCache.Redis.Set(ctx, key, value, ttl)
	if status.Err() != nil {
		return status.Err()
	}
	err := DefaultCache.Memory.Set(key, value, 10*time.Minute)
	if err != nil {
		return err
	}
	return nil
}

func Get(key string) ([]byte, error) {
	val, err := DefaultCache.Memory.Get(key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return val, nil
	}
	redisValue := DefaultCache.Redis.Get(ctx, key)
	if redisValue.Err() != nil {
		return nil, redisValue.Err()
	}
	bt, _ := redisValue.Bytes()
	DefaultCache.Memory.Set(key, bt, 10*time.Minute)
	return val, nil
}

func Delete(key string) error {
	status := DefaultCache.Redis.Del(ctx, key)
	if status.Err() != nil {
		return status.Err()
	}
	err := DefaultCache.Memory.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func Keys(pattern string) ([]string, error) {
	status := DefaultCache.Redis.Keys(ctx, pattern)
	if status.Err() != nil {
		return nil, status.Err()
	}
	keys, _ := status.Result()
	return keys, nil
}

func DeletePattern(key string) error {
	status := DefaultCache.Redis.Keys(ctx, key)
	if status.Err() != nil {
		return status.Err()
	}
	keys, _ := status.Result()
	st := DefaultCache.Redis.Del(ctx, keys...)
	if st.Err() != nil {
		return st.Err()
	}
	err := DefaultCache.Memory.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func Close() error {
	err := DefaultCache.Redis.Close()
	if err != nil {
		return err
	}
	err = DefaultCache.Memory.Close()
	if err != nil {
		return err
	}
	return nil
}

func Reset() error {
	status := DefaultCache.Redis.FlushDB(ctx)
	if status.Err() != nil {
		return status.Err()
	}
	err := DefaultCache.Memory.Reset()
	if err != nil {
		return err
	}
	return nil
}

func Client() *redis.Client {
	return DefaultCache.Redis
}
