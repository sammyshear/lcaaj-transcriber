package internal

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPhoneticAlphabet(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"A E I O U 3 1 6", "a e i o u ə ɪ ʌ"},
		{"B C D F G H J K L M N P R S T V W X Z", "b ts d f g h j k l m n p r s t v w x z"},
		{"95", "ʔ"},
		{"C+ C7 C2", "tʃ tʂ̻ ts̬"},
		{"S78 Z78 C78", "ʂ̻ʲ ʐ̻ʲ tʂ̻ʲ"},
		{"S+8,", "ʃʲ̩"},
		{"A+,,", "ˈã"},
		{"A495", "a̞."},
	} {
		t.Run(tc.in, func(t *testing.T) { assertTranscription(t, tc.in, tc.want) })
	}
}

func TestVowelDiacritics(t *testing.T) {
	for i, vowel := range []string{"A", "E", "I", "O", "U", "3", "1", "6"} {
		base := []string{"a", "e", "i", "o", "u", "ə", "ɪ", "ʌ"}[i]
		for _, mark := range []struct{ in, before, after string }{
			{".", "", "ː"},
			{"..", "", "ːː"},
			{"94", "", "̆"},
			{"95", "", "."},
			{",", "ˌ", ""},
			{",,", "ˈ", ""},
			{"+", "", "̃"},
			{"4", "", "̞"},
			{"5", "", "̝"},
			{"7", "", "̠"},
			{"8", "", "̟"},
		} {
			t.Run(vowel+mark.in, func(t *testing.T) { assertTranscription(t, vowel+mark.in, mark.before+base+mark.after) })
		}
	}
}

func TestConsonantDiacritics(t *testing.T) {
	for _, rule := range []struct{ letters, suffix, mark string }{
		{"bdgjlmnrvwz", "2", "̥"},
		{"cfhkpstx", "2", "̬"},
		{"bdgjlmnrvwfhkptx", "7", "ˠ"},
		{"bdgjlmnrvwfhkptxscz", "8", "ʲ"},
		{"bdfgkptv", "+", "ⁿ"},
		{"bcdfghjklmnprstvwxz", ",", "̩"},
	} {
		for _, letter := range rule.letters {
			base := string(letter)
			if base == "c" {
				base = "ts"
			}
			t.Run(string(letter)+rule.suffix, func(t *testing.T) {
				assertTranscription(t, strings.ToUpper(string(letter))+rule.suffix, base+rule.mark)
			})
		}
	}
}

func TestMixedNotation(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", ""},
		{"A+ QADJ", "ã adjective"},
		{"A94 QBF", "ă yes, fragment in book"},
		{"A QENG Case, C+ 95. QP S+", "a explanation in english:  Case, C+ 95.  ʃ"},
		{"QENG FirstQP QENG SecondQP", "explanation in english:  First explanation in english:  Second"},
		{"QGLE KIND OF PASSOVER DISH QP KI.X3L3X", "informant's explanation in English:  KIND OF PASSOVER DISH  kiːxələx"},
		{"QGLY C+ 95", "informant's explanation in Yiddish:  C+ 95"},
		{"QEDN QEDS QINF QNOUN QNN QN QSMT", "editor disagrees editor's query infinitive noun very new new, modern sometimes"},
		{"($ (( )+ )- )=", " relevant to another geographic location  reference to dictionary  prompted and accepted  prompted and rejected  prompted and replaces preceding response"},
		{"+ BUT Hello, World!QP A", " yes but:  Hello, World! a"},
		{`QENG \, \: 100% QP`, `explanation in english:  \, \: 100% `},
		{"QENG Unicode: ייִדיש!", "explanation in english:  Unicode: ייִדיש!"},
		{"MISTD", " misunderstanding, informant's response does not apply to question (editor's comment)"},
	} {
		t.Run(tc.in, func(t *testing.T) { assertTranscription(t, tc.in, tc.want) })
	}
}

func assertTranscription(t *testing.T, input, want string) {
	t.Helper()
	if got := Transcribe(&dataSignal{Data: input}); got != want {
		t.Errorf("Transcribe(%q)\n got: %q (%U)\nwant: %q (%U)", input, got, []rune(got), want, []rune(want))
	}
}

func FuzzTranscribe(f *testing.F) {
	for _, seed := range []string{"A94 QADJ", "S78,", "QENG Hello!QP C+", "QED oneQP QED twoQP", "($", "", "ייִדיש"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if !utf8.ValidString(input) {
			t.Skip()
		}
		got := Transcribe(&dataSignal{Data: input})
		if !utf8.ValidString(got) {
			t.Fatalf("invalid UTF-8 for %q", input)
		}
		if again := Transcribe(&dataSignal{Data: input}); got != again {
			t.Fatalf("nondeterministic result for %q", input)
		}
	})
}

// The fixture is hand-authored: expected results must not be generated from the
// production rule tables, since that would reproduce the same mistakes.
func TestNotationKey(t *testing.T) {
	fixture, err := os.ReadFile("testdata/notation.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(strings.TrimSuffix(string(fixture), "\n"), "\n") {
		input, want, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatalf("invalid fixture row %q", line)
		}
		t.Run(input, func(t *testing.T) { assertTranscription(t, input, want) })
	}
}

// Independent calls must be safe for simultaneous HTTP requests.
func TestConcurrentTranscription(t *testing.T) {
	for i := range 16 {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			assertTranscription(t, "C+ A94 QENG Case!QP S78", "tʃ ă explanation in english:  Case! ʂ̻ʲ")
		})
	}
}
