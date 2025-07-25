package common

import (
	"sync"

	"github.com/pemistahl/lingua-go"
)

var detector lingua.LanguageDetector
var once sync.Once

func GetDetector() lingua.LanguageDetector {
	once.Do(func() {
		languages := []lingua.Language{
			lingua.English,
			lingua.French,
			lingua.German,
			lingua.Spanish,
			lingua.Chinese,
		}

		detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			Build()
	})
	return detector
}

func GetLanguage(content string) string {
	longShortLanguageMap := map[string]string{
		"English":           "en",
		"Chinese":           "zh-cn",
		"Afrikaans":         "af",
		"Albanian":          "sq",
		"Arabic":            "ar",
		"Armenian":          "hy",
		"Azerbaijani":       "az",
		"Basque":            "eu",
		"Belarusian":        "be",
		"Bengali":           "bn",
		"Norwegian Bokmal":  "nb",
		"Bosnian":           "bs",
		"Bulgarian":         "bg",
		"Catalan":           "ca",
		"Croatian":          "hr",
		"Czech":             "cs",
		"Danish":            "da",
		"Dutch":             "nl",
		"Esperanto":         "eo",
		"Estonian":          "et",
		"Finnish":           "fi",
		"French":            "fr",
		"Ganda":             "lg",
		"Georgian":          "ka",
		"German":            "de",
		"Greek":             "el",
		"Gujarati":          "gu",
		"Hebrew":            "he",
		"Hindi":             "hi",
		"Hungarian":         "hu",
		"Icelandic":         "is",
		"Indonesian":        "id",
		"Irish":             "ga",
		"Italian":           "it",
		"Japanese":          "ja",
		"Kazakh":            "kk",
		"Korean":            "ko",
		"Latin":             "la",
		"Latvian":           "lv",
		"Lithuanian":        "lt",
		"Macedonian":        "mk",
		"Malay":             "ms",
		"Maori":             "mi",
		"Marathi":           "mr",
		"Mongolian":         "mn",
		"Norwegian Nynorsk": "nn",
		"Persian":           "fa",
		"Polish":            "pl",
		"Portuguese":        "pt",
		"Punjabi":           "pa",
		"Romanian":          "ro",
		"Russian":           "ru",
		"Serbian":           "sr",
		"Shona":             "sn",
		"Slovak":            "sk",
		"Slovene":           "sl",
		"Somali":            "so",
		"Sotho":             "st",
		"Spanish":           "es",
		"Swahili":           "sw",
		"Swedish":           "sv",
		"Tagalog":           "tl",
		"Tamil":             "ta",
		"Telugu":            "te",
		"Thai":              "th",
		"Tsonga":            "ts",
		"Tswana":            "tn",
		"Turkish":           "tr",
		"Ukrainian":         "uk",
		"Urdu":              "ur",
		"Vietnamese":        "vi",
		"Welsh":             "cy",
		"Xhosa":             "xh",
		"Yoruba":            "yo",
		"Zulu":              "zu",
	}

	//detector := lingua.NewLanguageDetectorBuilder().FromAllLanguages().Build()

	if language, exists := GetDetector().DetectLanguageOf(content); exists {
		l, ok := longShortLanguageMap[language.String()]
		if ok {
			return l
		}
		return language.String()
	}

	return "other"
}
