//go:build !structured

package internal

import (
	"fmt"
	"regexp"
	"strings"
)

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
	Vowels                 = "([aeiouəɪʌ])"
	VowelsFull             = "([aeiouəɪʌ][\\x{0306}\\x{0303}\\x{031E}\\x{031D}\\x{0320}\\x{031F}]?)"
	Consonants             = "((?:tʂ̻|tʃ|ʂ̻|ʐ̻|[ʔbcdfghjklmnprstvwxzʃʒ])[\\x{207F}\\x{02B2}\\x{02E0}]*)"
	HushedConsonants       = "([csz])"
	NasalReleaseConsonants = "([bdfgkptv])"
	UnvoicingConsonants    = "([bdgjlmnrvwz])"
	VoicingConsonants      = "([cfhkpstx])"
	VelarizingConsonants   = "([bdgjlmnrvwfhkptx])"
	PalatalizingConsonants = "(tʂ̻|tʃ|ʂ̻|ʐ̻|[bdgjlmnrvwfhkptxsczʃʒ])"
	NonPhoneticNotation    = "%s(?<text>(?s:.*?))(?:QP|$)"
	BaseNotation           = "(%s)"
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
	{regexp.MustCompile(Vowels + "(94)"), "%c\u0306"},
	{regexp.MustCompile(Vowels + "(\\+)"), "%c\u0303"},
	{regexp.MustCompile(Vowels + "(4)"), "%c\u031E"},
	{regexp.MustCompile(Vowels + "(5)"), "%c\u031D"},
	{regexp.MustCompile(Vowels + "(7)"), "%c\u0320"},
	{regexp.MustCompile(Vowels + "(8)"), "%c\u031F"},
	{regexp.MustCompile(VowelsFull + "(95)"), "%c."},
	{regexp.MustCompile(VowelsFull + "(,)(,)"), "\u02C8%c"},
	{regexp.MustCompile(VowelsFull + "(,)"), "\u02CC%c"},
}

var consKeys = []RegexpKey{
	{regexp.MustCompile(HushedConsonants + "(\\+)"), "hushed"},
	{regexp.MustCompile(HushedConsonants + "(7)"), "semi-hushed"},
	{regexp.MustCompile(UnvoicingConsonants + "(2)"), "%c\u0325"},
	{regexp.MustCompile(VoicingConsonants + "(2)"), "%c\u032C"},
	{regexp.MustCompile(VelarizingConsonants + "(7)"), "%c\u02E0"},
	{regexp.MustCompile(PalatalizingConsonants + "(8)"), "%c\u02B2"},
	{regexp.MustCompile(NasalReleaseConsonants + "(\\+)"), "%c\u207F"},
	{regexp.MustCompile(Consonants + "(,)"), "%c\u0329"},
}

// keys for notations

var notKeys = []RegexpKey{
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "0")), "question not asked"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, `(?:^|[^A-Za-z\d)])(\+ BUT)`)), " yes but: %s"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, `(?:^|[^A-Za-z\d)])(\- BUT)`)), " no but: %s"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\+\$)`), " yes\\, but doubtful"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\-\$)`), " no\\, but doubtful"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\+)`), " yes"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\-)`), " no"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\=)`), " self-corrected"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\#)`), " self-corrected"},
	{regexp.MustCompile(`(?:^|[^A-Za-z\d)])(\*)`), " QFQM"},
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
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(ADJ)")), "adjective"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(AMER)")), "american yiddish development"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(ANG)")), "anglicism"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(AP)")), "applies to"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, `Q(\-AP)`)), "does not apply to"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(BF)")), "yes, fragment in book"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(B)")), "yes, text in protocol book"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(CF)")), "interviewer's comment: compare"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(DG)")), "disgust"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(EDS)")), "editor's query"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(EDN)")), "editor disagrees"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(ED)")), "editor's comments follow: %s"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(ELSW)")), "elsewhere"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(EM)")), "emphatic"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(ENG)")), "explanation in english: %s"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(ETC)")), "etc\\:"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(ET)")), "etymology supplied by informant"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(FR)")), "yes, fragment on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, `Q(F\/Y)`)), "response of wife or other female bystander"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(GERM)")), "Informant’s statement that word is not Yiddish but German"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(GLE)")), "informant's explanation in English: %s"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(GLY)")), "informant's explanation in Yiddish: %s"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(GL)")), "gloss"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(HUM)")), "amusing"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(HUNG)")), "informant's statement that word is not Yiddish but Hungarian"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(H)")), "heard but not used"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(INF)")), "infinitive"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(I GL)")), "Interviewer's Summary: %s"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(I)")), "interviewer's comments follow: %s"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(K)")), "known"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, `Q(\-K)`)), "unknown"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(LAT)")), "not on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(LIT)")), "literary"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(MEMX)")), "informant's surprise at own recollection"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, `Q(M\/Y)`)), "response by husband or other male bystander"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(NEX)")), "did not exist"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(N)")), "new, modern"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(NN)")), "very new"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(NOUN)")), "noun"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(NP)")), "unprompted answer to prompted question"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(NT)")), "not on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(OF)")), "oldfashioned"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(OOF)")), "Very Oldfashioned"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(OTW)")), "Otherwise"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(P)")), ""},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(POL)")), "Informant's statement that word is not Yiddish but Polish"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(Q)")), "Check answer on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(RR)")), "very rare"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(RUM)")), "Informant's statement that word is not Yiddish but Rumanian"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(RUS)")), "Informant's statement that word is not Yiddish but Russian"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(RTR)")), "rather"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(R)")), "rare"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(SMT)")), "sometimes"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(SYN)")), "synonym"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(S)")), "said by: %s"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(TA)")), "tape audited"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(TF)")), "yes, fragment on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(T)")), "yes, text on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, `Q(\-T)`)), "text not on tape"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(UU)")), "VeryCommon"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(U)")), "Usual"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, `Q(\-U)`)), "Unusual"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(VB)")), "Verb"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(VL)")), "Vulgar"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(V)")), "Proverb"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(W)")), "used by: %s"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, `Q(\-W)`)), "not used by: %s"},
	{regexp.MustCompile(fmt.Sprintf(NonPhoneticNotation, "Q(YID)")), "Informant's explanation in Yiddish but not necessarily verbatim or phonetically accurate: %s"},
	{regexp.MustCompile(fmt.Sprintf(BaseNotation, "Q(ZZ)")), "interviewer's comment: not elicitable"},
}

// Transcribe separates source notation from phonetic spans before rendering either.
// Generated notes are never fed back through the transcription rules.
func Transcribe(data *dataSignal) string {
	var out strings.Builder
	source := data.Data
	for len(source) > 0 {
		var rule *RegexpKey
		var span []int
		markerLength := 0
		for i := range notKeys {
			candidate := notKeys[i].FindStringSubmatchIndex(source)
			length := 0
			if candidate != nil {
				length = candidate[1] - candidate[0]
				if text := notKeys[i].SubexpIndex("text"); text >= 0 {
					length = candidate[2*text] - candidate[0]
				}
			}
			if candidate != nil && (span == nil || candidate[0] < span[0] ||
				(candidate[0] == span[0] && length > markerLength)) {
				rule, span, markerLength = &notKeys[i], candidate, length
			}
		}
		if span == nil {
			out.WriteString(transcribePhonetics(source))
			break
		}
		out.WriteString(transcribePhonetics(source[:span[0]]))
		matched := source[span[0]:span[1]]
		out.WriteString(transcribeNotation(*rule, matched))
		source = source[span[1]:]
	}
	return out.String()
}

// transcribeNotation renders only the matched source notation. Template escaping
// is removed before interpolation so literal comment text remains unchanged.
func transcribeNotation(rule RegexpKey, matched string) string {
	note := strings.ReplaceAll(rule.name, "\\,", ",")
	note = strings.ReplaceAll(note, "\\:", ".")
	if index := rule.SubexpIndex("text"); index >= 0 {
		note = fmt.Sprintf(note, rule.FindStringSubmatch(matched)[index])
	}
	return note
}

func transcribePhonetics(source string) string {
	o := strings.ToLower(source)
	for k, v := range basicMap {
		o = strings.ReplaceAll(o, k, v)
	}

	for _, k := range vowelKeys {
		o = k.ReplaceAllStringFunc(o, func(match string) string {
			base := k.FindStringSubmatch(match)[1]
			return fmt.Sprintf(strings.ReplaceAll(k.name, "%c", "%s"), base)
		})
	}
	for _, k := range consKeys {
		o = k.ReplaceAllStringFunc(o, func(match string) string {
			base := k.FindStringSubmatch(match)[1]
			switch k.name {
			case "hushed":
				return hushedMap[base]
			case "semi-hushed":
				return semiHushedMap[base]
			default:
				return fmt.Sprintf(strings.ReplaceAll(k.name, "%c", "%s"), base)
			}
		})
	}

	o = strings.ReplaceAll(o, "c", "ts")
	o = strings.ReplaceAll(o, "95", "ʔ")
	o = strings.ReplaceAll(o, "\\,", ",")
	o = strings.ReplaceAll(o, "\\:", ".")
	return o
}
