package storage

import (
	"fmt"

	"github.com/spf13/viper"
)

type MemoryStorageCfg struct {
	LocaleFiles []string `mapstructure:"locale_files" json:"locale_files"`
}

type memoryStorage struct {
	cfg   *MemoryStorageCfg
	viper *viper.Viper
}

func NewMemoryStorage(cfg *MemoryStorageCfg) Storage {
	memoryStorage := memoryStorage{
		cfg:   cfg,
		viper: viper.New(),
	}
	memoryStorage.viper.SetConfigName("locale")
	for _, filepath := range cfg.LocaleFiles {
		viper.AddConfigPath(filepath)
		viper.MergeInConfig()
	}
	return &memoryStorage
}

func (cache *memoryStorage) GetLocalizedMessage(locale string, key string) string {
	return cache.viper.GetString(fmt.Sprintf("%s.%s", locale, key))
}
