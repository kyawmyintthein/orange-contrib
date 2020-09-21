package storage

import (
	"fmt"

	"github.com/spf13/viper"
)

type memoryCache struct {
	viper *viper.Viper
}

func NewMemoryCache() Storage {
	memoryCache := memoryCache{
		viper: viper.New(),
	}
	return &memoryCache
}

func (cache *memoryCache) GetLocalizedMessage(locale string, key string) string {
	return cache.viper.GetString(fmt.Sprintf("%s.%s", locale, key))
}
