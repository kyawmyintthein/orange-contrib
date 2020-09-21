package localex

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyawmyintthein/orange-contrib/localex/storage"
)

const (
	X_LOCALE = "X-LOCALE"
)

type TranslatorCfg struct {
	Enabled       bool   `mapstructure:"enabled" json:"enabled"`
	DefaultLocale string `mapstructure:"default_locale" json:"default_locale"`
}

type Translator interface {
	Translate(context.Context, string, ...string) string
}

type translator struct {
	cfg     *TranslatorCfg
	storage storage.Storage
}

func NewTranslator(cfg *TranslatorCfg, storage storage.Storage) Translator {
	return &translator{
		cfg:     cfg,
		storage: storage,
	}
}

func (t *translator) Translate(ctx context.Context, messageID string, argKvs ...string) string {

	translatedString := messageID

	if t.cfg.Enabled {

		locale, _ := ctx.Value(X_LOCALE).(string)
		if locale == "" {
			locale = t.cfg.DefaultLocale
		}
		localizedString := t.storage.GetLocalizedMessage(messageID, locale)
		if localizedString != "" {
			translatedString = localizedString
		}
	}

	if len(argKvs) != 0 {
		argsMap := make(map[string]string)
		previousKey := ""
		for _, v := range argKvs {
			if previousKey != "" {
				argsMap[previousKey] = v
			}
			previousKey = v
		}

		for k, v := range argsMap {
			translatedString = strings.Replace(translatedString, fmt.Sprintf("{{var_%s}}", k), v, -1)
		}

	}
	return translatedString
}
