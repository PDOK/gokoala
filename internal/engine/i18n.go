package engine

import (
	"github.com/BurntSushi/toml"
	"github.com/PDOK/gokoala/config"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func newLocalizers(availableLanguages []config.Language) map[language.Tag]i18n.Localizer {
	localizers := make(map[language.Tag]i18n.Localizer)
	// add localizer for each available language
	for _, lang := range availableLanguages {
		bundle := i18n.NewBundle(lang.Tag)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		bundle.MustLoadMessageFile("assets/i18n/active." + lang.String() + ".toml")
		localizers[lang.Tag] = *i18n.NewLocalizer(bundle, lang.String())
	}
	return localizers
}
