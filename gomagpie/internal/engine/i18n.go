package engine

import (
	"github.com/PDOK/gomagpie/config"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

func newLocalizers(availableLanguages []config.Language) map[language.Tag]i18n.Localizer {
	localizers := make(map[language.Tag]i18n.Localizer)
	// add localizer for each available language
	for _, lang := range availableLanguages {
		bundle := i18n.NewBundle(lang.Tag)
		bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
		bundle.MustLoadMessageFile("assets/i18n/" + lang.String() + ".yaml")
		localizers[lang.Tag] = *i18n.NewLocalizer(bundle, lang.String())
	}
	return localizers
}
