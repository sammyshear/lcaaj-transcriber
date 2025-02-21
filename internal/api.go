package internal

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/sammyshear/lcaaj-transcriber/views"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type dataSignal struct {
	Data string `json:"data"`
}

type RegexpKey struct {
	*regexp.Regexp
	name string
}

var basicMap = map[string]string{
	"3": "ə",
	"1": "ɪ",
	"6": "ʌ",
	".": "ː",
}

const (
	VOWELS                   = "([aeiouəɪʌ])"
	VOWELS_FULL              = "([aeiouəɪʌ][\\x{0306}\\x{0303}\\x{031E}\\x{031D}\\x{0320}\\x{031F}]?)"
	CONSONANTS               = "([ʔbcdfghjklmnprstvwxzʃʒ(tʃ)ʂ̻ʐ̻(tʂ̻)][\\x{207F}\\x{02B2}\\x{02E0}]?)"
	HUSHED_CONSONANTS        = "([csz])"
	NASAL_RELEASE_CONSONANTS = "([bdfgkptv])"
	UNVOICING_CONSONANTS     = "([bdgjlmnrvwz])"
	VOICING_CONSONANTS       = "([cfhkpstx])"
	VELARIZING_CONSONANTS    = "([bdgjlmnrvwfhkptx])"
	PALATALIZING_CONSONANTS  = "([bdgjlmnrvwfhkptxsczʃʒ(tʃ)ʂ̻ʐ̻(tʂ̻)])"
)

var hushedMap = map[string]string{
	"s": "ʃ",
	"z": "ʒ",
	"c": "tʃ",
}

var semiHushedMap = map[string]string{
	"s": "ʂ̻",
	"c": "tʂ̻",
	"z": "ʐ̻",
}

var vowelKeys = []RegexpKey{
	{regexp.MustCompile(VOWELS + "(94)"), "diacShort"},
	{regexp.MustCompile(VOWELS + "(\\+)"), "diacNasal"},
	{regexp.MustCompile(VOWELS + "(4)"), "diacLower"},
	{regexp.MustCompile(VOWELS + "(5)"), "diacRaise"},
	{regexp.MustCompile(VOWELS + "(7)"), "diacBack"},
	{regexp.MustCompile(VOWELS + "(8)"), "diacFront"},
	{regexp.MustCompile(VOWELS_FULL + "(95)"), "diacSyllabEnd"},
	{regexp.MustCompile(VOWELS_FULL + "(,)(,)"), "diacStress"},
	{regexp.MustCompile(VOWELS_FULL + "(,)"), "diacSecondaryStress"},
}

var consKeys = []RegexpKey{
	{regexp.MustCompile(HUSHED_CONSONANTS + "(\\+)"), "diacHushing"},
	{regexp.MustCompile(HUSHED_CONSONANTS + "(7)"), "diacSemiHushing"},
	{regexp.MustCompile(UNVOICING_CONSONANTS + "(2)"), "diacUnvoicing"},
	{regexp.MustCompile(VOICING_CONSONANTS + "(2)"), "diacVoicing"},
	{regexp.MustCompile(VELARIZING_CONSONANTS + "(7)"), "diacVelarizing"},
	{regexp.MustCompile(PALATALIZING_CONSONANTS + "(8)"), "diacPalatalized"},
	{regexp.MustCompile(NASAL_RELEASE_CONSONANTS + "(\\+)"), "diacNasalRelease"},
	{regexp.MustCompile(CONSONANTS + "(,)"), "diacSyllab"},
}

var vowelsMap = map[string]string{
	vowelKeys[0].name: "%c\u0306",
	vowelKeys[1].name: "%c\u0303",
	vowelKeys[2].name: "%c\u031E",
	vowelKeys[3].name: "%c\u031D",
	vowelKeys[4].name: "%c\u0320",
	vowelKeys[5].name: "%c\u031F",
	vowelKeys[6].name: "%c.",
	vowelKeys[7].name: "\u02C8%c",
	vowelKeys[8].name: "\u02CC%c",
}

var consMap = map[string]string{
	consKeys[0].name: "hushed",
	consKeys[1].name: "semi-hushed",
	consKeys[2].name: "%c\u0325",
	consKeys[3].name: "%c\u032C",
	consKeys[4].name: "%c\u02E0",
	consKeys[5].name: "%c\u02B2",
	consKeys[6].name: "%c\u207F",
	consKeys[7].name: "%c\u0329",
}

func Transcribe(w http.ResponseWriter, r *http.Request) {
	data := &dataSignal{}
	err := datastar.ReadSignals(r, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sse := datastar.NewSSE(w, r)
	o := strings.ToLower(data.Data)

	for k, v := range basicMap {
		o = strings.ReplaceAll(o, k, v)
	}

	for _, k := range vowelKeys {
		v := vowelsMap[k.name]
		o = k.ReplaceAllStringFunc(o, func(s string) string {
			r, _ := utf8.DecodeRuneInString(s)
			_, lastSize := utf8.DecodeLastRuneInString(s)
			woLastRune := s[:len(s)-lastSize]
			if c, size := utf8.DecodeLastRuneInString(woLastRune); c != r && size != 0 && c != ',' && c != '9' {
				return fmt.Sprintf(v+"%c", r, c)
			} else if c == ',' {
				woLastRune = woLastRune[:len(woLastRune)-size]
				if cr, size := utf8.DecodeLastRuneInString(woLastRune); cr != r && size != 0 {
					return fmt.Sprintf(v+"%c", r, cr)
				}
			} else if c == '9' {
				woLastRune = woLastRune[:len(woLastRune)-size]
				if cr, size := utf8.DecodeLastRuneInString(woLastRune); cr != r && size != 0 {
					return fmt.Sprintf(v+"%c", r, cr)
				}
			}
			return fmt.Sprintf(v, r)
		})
	}

	for _, k := range consKeys {
		v := consMap[k.name]
		o = k.ReplaceAllStringFunc(o, func(s string) string {
			r, _ := utf8.DecodeRuneInString(s)
			switch v {
			case "hushed":
				for key, val := range hushedMap {
					if string(r) == key {
						return val
					}
				}
			case "semi-hushed":
				for key, val := range semiHushedMap {
					if string(r) == key {
						return val
					}
				}
			}
			_, lastSize := utf8.DecodeLastRuneInString(s)
			woLastRune := s[:len(s)-lastSize]
			if c, size := utf8.DecodeLastRuneInString(woLastRune); c != r && size != 0 {
				return fmt.Sprintf(v+"%c", r, c)
			}
			return fmt.Sprintf(v, r)
		})

		o := strings.ReplaceAll(o, "95", "ʔ")

		sse.MergeFragmentTempl(views.Transcription(o), datastar.WithSelectorID("result"))
	}
}
