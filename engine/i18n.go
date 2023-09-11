package engine

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func newLocalizers(availableLanguages []language.Tag) map[language.Tag]i18n.Localizer {
	localizers := make(map[language.Tag]i18n.Localizer)
	// add localizer for each available language
	for _, lang := range availableLanguages {
		bundle := i18n.NewBundle(lang)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		bundle.MustLoadMessageFile("assets/i18n/active." + lang.String() + ".toml")
		localizers[lang] = *i18n.NewLocalizer(bundle, lang.String())
	}
	return localizers
}
