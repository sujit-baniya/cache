package cache

import (
	"github.com/gofiber/storage/memory"
	"github.com/gofiber/storage/redis"
	"time"
)

type Config struct {
	Host string
	Username string
	Password string
	DB int
	Port int
}

type Cache struct {
	Memory *memory.Storage
	Redis *redis.Storage
}

var DefaultCache = &Cache{
	Memory: memory.New(),
	Redis: redis.New(),
}

func New(cfg ...Config) *Cache {
	cs := &Cache{Memory: memory.New()}
	if len(cfg) == 0 {
		return DefaultCache
	}
	config := cfg[0]
	cs.Redis = redis.New(redis.Config{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		Database: config.DB,
	})
	DefaultCache = cs
	return cs
}

func Set(key string, value []byte, ttl time.Duration) error {
	err := DefaultCache.Redis.Set(key, value, ttl)
	if err != nil {
		return err
	}
	err = DefaultCache.Memory.Set(key, value, 10 * time.Minute)
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
	val, err = DefaultCache.Redis.Get(key)
	if err != nil {
		return nil, err
	}
	DefaultCache.Memory.Set(key, val, 10 * time.Minute)
	return val, nil
}

func Delete(key string) error {
	err := DefaultCache.Redis.Delete(key)
	if err != nil {
		return err
	}
	err = DefaultCache.Memory.Delete(key)
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
	err := DefaultCache.Redis.Reset()
	if err != nil {
		return err
	}
	err = DefaultCache.Memory.Reset()
	if err != nil {
		return err
	}
	return nil
}