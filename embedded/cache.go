package embedded

import (
	"github.com/gofiber/storage/badger"
	"github.com/gofiber/storage/memory"
	"time"
)

type Config struct {
	DB string
}

type Cache struct {
	Memory *memory.Storage
	DB     *badger.Storage
}

var DefaultCache = &Cache{
	Memory: memory.New(),
	DB:     badger.New(),
}

func New(cfg ...Config) *Cache {
	cs := &Cache{Memory: memory.New()}
	if len(cfg) == 0 {
		return DefaultCache
	}
	config := cfg[0]
	cs.DB = badger.New(badger.Config{
		Database: config.DB,
	})
	DefaultCache = cs
	return cs
}

func Set(key string, value []byte, ttl time.Duration) error {
	err := DefaultCache.DB.Set(key, value, ttl)
	if err != nil {
		return err
	}
	err = DefaultCache.Memory.Set(key, value, 5*time.Minute)
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
	val, err = DefaultCache.DB.Get(key)
	if err != nil {
		return nil, err
	}
	DefaultCache.Memory.Set(key, val, 5*time.Minute)
	return val, nil
}

func Delete(key string) error {
	err := DefaultCache.DB.Delete(key)
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
	err := DefaultCache.DB.Close()
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
	err := DefaultCache.DB.Reset()
	if err != nil {
		return err
	}
	err = DefaultCache.Memory.Reset()
	if err != nil {
		return err
	}
	return nil
}
