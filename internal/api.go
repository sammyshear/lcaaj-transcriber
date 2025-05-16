package internal

import (
	"encoding/json"
	"fmt"
	"io"
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
	NON_PHONETIC_NOTATION    = "%s(?<text>[A-Za-z\\d\\s]*)(QP)"
	BASE_NOTATION            = "(%s)"
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
	{regexp.MustCompile(VOWELS + "(94)"), "%c\u0306"},
	{regexp.MustCompile(VOWELS + "(\\+)"), "%c\u0303"},
	{regexp.MustCompile(VOWELS + "(4)"), "%c\u031E"},
	{regexp.MustCompile(VOWELS + "(5)"), "%c\u031D"},
	{regexp.MustCompile(VOWELS + "(7)"), "%c\u0320"},
	{regexp.MustCompile(VOWELS + "(8)"), "%c\u031F"},
	{regexp.MustCompile(VOWELS_FULL + "(95)"), "%c."},
	{regexp.MustCompile(VOWELS_FULL + "(,)(,)"), "\u02C8%c"},
	{regexp.MustCompile(VOWELS_FULL + "(,)"), "\u02CC%c"},
}

var consKeys = []RegexpKey{
	{regexp.MustCompile(HUSHED_CONSONANTS + "(\\+)"), "hushed"},
	{regexp.MustCompile(HUSHED_CONSONANTS + "(7)"), "semi-hushed"},
	{regexp.MustCompile(UNVOICING_CONSONANTS + "(2)"), "%c\u0325"},
	{regexp.MustCompile(VOICING_CONSONANTS + "(2)"), "%c\u032C"},
	{regexp.MustCompile(VELARIZING_CONSONANTS + "(7)"), "%c\u02E0"},
	{regexp.MustCompile(PALATALIZING_CONSONANTS + "(8)"), "%c\u02B2"},
	{regexp.MustCompile(NASAL_RELEASE_CONSONANTS + "(\\+)"), "%c\u207F"},
	{regexp.MustCompile(CONSONANTS + "(,)"), "%c\u0329"},
}

// keys for notations

var notKeys = []RegexpKey{
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "0")), "question not asked"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, `(?:^|[^A-Za-z\d])(\+ BUT)`)), " yes but: %s"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, `(?:^|[^A-Za-z\d])(\- BUT)`)), " no but: %s"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\+\$)`), " yes\\, but doubtful"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\-\$)`), " no\\, but doubtful"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\+)`), " yes"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\-)`), " no"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\=)`), " self-corrected"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\#)`), " self-corrected"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d])(\*)`), " QFQM"},
	{regexp.MustCompile(`\$`), " query"},
	{regexp.MustCompile(`(\|\|)`), " is different from"},
	{regexp.MustCompile(`(\/\/)(?<text>[A-Za-z\d]*)`), "(%s)"},
	{regexp.MustCompile(`(\)\+)`), " prompted and accepted"},
	{regexp.MustCompile(`(\)\-)`), " prompted and rejected"},
	{regexp.MustCompile(`(\)\=)`), " prompted and replaces preceding response"},
	{regexp.MustCompile(`(\(\/)`), " relevant to another question number"},
	{regexp.MustCompile(`(\(\$)`), " relevant to another geographic location"},
	{regexp.MustCompile(`(\(\()`), " reference to dictionary"},
	{regexp.MustCompile(`(\()`), " relevant to problem number in dialectology"},
	{regexp.MustCompile("(CLN)"), ":"},
	{regexp.MustCompile("(CM)"), "\\,"},
	{regexp.MustCompile("(DRWG)"), "drawing in protocol book"},
	{regexp.MustCompile("(EQ)"), " is identical with (in respect to some significant point)"},
	{regexp.MustCompile("(MISPMP)"), " misprompted (editor's comment)"},
	{regexp.MustCompile("(MISTD)"), " misunderstanding\\, informant's response does not apply to question (editor's comment)"},
	{regexp.MustCompile("(OVRPMP)"), " overprompted (editor's comment)"},
	{regexp.MustCompile("(SC)"), ";"},
	{regexp.MustCompile("(XX)"), " (sic)"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(ADJ)")), "adjective"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(AMER)")), "american yiddish development"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(ANG)")), "anglicism"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(AP)")), "applies to"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, `Q(\-AP)`)), "does not apply to"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(BF)")), "yes, fragment in book"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(B)")), "yes, text in protocol book"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(CF)")), "interviewer's comment: compare"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(DG)")), "disgust"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(EDS)")), "editor's query"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(EDN)")), "editor disagrees"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(ED)")), "editor's comments follow: %s"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(ELSW)")), "elsewhere"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(EM)")), "emphatic"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(ENG)")), "explanation in english: %s"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(ETC)")), "etc\\:"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(ET)")), "etymology supplied by informant"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(FR)")), "yes, fragment on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, `Q(F\/Y)`)), "response of wife or other female bystander"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(GERM)")), "Informant’s statement that word is not Yiddish but German"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(GLE)")), "informant's explanation in English: "},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(GLY)")), "informant's explanation in Yiddish: "},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(GL)")), "gloss"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(HUM)")), "amusing"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(HUNG)")), "informant's statement that word is not Ydidish but Hungarian"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(H)")), "heard but not used"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(INF)")), "infinitive"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(I GL)")), "Interviewer's Summary: %s"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(I)")), "interviewer's comments follow: %s"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(K)")), "known"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, `Q(\-K)`)), "unknown"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(LAT)")), "not on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(LIT)")), "literary"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(MEMX)")), "informant's surpise at own recollection"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, `Q(M\/Y)`)), "response by husband or other male bystander"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(NEX)")), "did not exist"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(NN)")), "notVeryNew"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(NOUN)")), "noun"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(NP)")), "unprompted answer to prompted question"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(NT)")), "not on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(OF)")), "oldfashioned"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(OOF)")), "Very Oldfashioned"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(OTW)")), "Otherwise"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(POL)")), "Informant's statement that word is not Yiddish but Polish"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(Q)")), "Check answer on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(RR)")), "very rare"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(RUM)")), "Informant's statement that word is not Yiddish but Rumanian"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(RUS)")), "Informant's statement that word is not Yiddish but Russian"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(RTR)")), "rather"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(R)")), "rare"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(SMT)")), "notSometimes"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(SYN)")), "synonym"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(S)")), "said by: %s"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(TA)")), "tape audited"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(TF)")), "yes, fragment on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(T)")), "yes, text on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, `Q(\-T)`)), "text not on tape"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(UU)")), "VeryCommon"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(U)")), "Usual"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, `Q(\-U)`)), "Unusual"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(VB)")), "Verb"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(VL)")), "Vulgar"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(V)")), "Proverb"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(W)")), "used by: %s"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, `Q(\-W)`)), "not used by: %s"},
	{regexp.MustCompile(fmt.Sprintf(NON_PHONETIC_NOTATION, "Q(YID)")), "Informant's explanation in Yiddish but not necessarily verbatim or phoenetically accurate: %s"},
	{regexp.MustCompile(fmt.Sprintf(BASE_NOTATION, "Q(ZZ)")), "interviewer's comment: not elicitable"},
}

func Transcribe(data *dataSignal) string {
	o := data.Data

	for _, k := range notKeys {
		o = k.ReplaceAllStringFunc(o, func(s string) string {
			ans := k.name
			if strings.HasSuffix(ans, "%s") || strings.HasSuffix(ans, "(%s)") {
				groupIndex := k.SubexpIndex("text")
				matches := k.FindStringSubmatch(o)
				return fmt.Sprintf(ans, matches[groupIndex])
			}

			return k.name
		})
	}

	o = strings.ToLower(o)

	for k, v := range basicMap {
		o = strings.ReplaceAll(o, k, v)
	}

	for _, k := range vowelKeys {
		v := k.name
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
		v := k.name
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

		o = strings.ReplaceAll(o, "95", "ʔ")
		o = strings.ReplaceAll(o, "c", "ts")
	}

	o = strings.ReplaceAll(o, "\\,", ",")
	o = strings.ReplaceAll(o, "\\:", ".")
	return o
}

func ApiTranscribe(w http.ResponseWriter, r *http.Request) {
	data := &dataSignal{}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	json.Unmarshal(b, data)
	w.Write([]byte(Transcribe(data)))
}

func DatastarTranscribe(w http.ResponseWriter, r *http.Request) {
	data := &dataSignal{}
	err := datastar.ReadSignals(r, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sse := datastar.NewSSE(w, r)
	o := Transcribe(data)
	sse.MergeFragmentTempl(views.Transcription(o), datastar.WithSelectorID("result"))
}
