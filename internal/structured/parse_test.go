package structured

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSoundTokens(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   Sound
	}{
		{"S78", Sound{"S", []Modifier{SemiHushed, Palatalized}}},
		{"A+,,", Sound{"A", []Modifier{Nasalized, PrimaryStress}}},
		{"C2", Sound{"C", []Modifier{Voiced}}},
		{"A495", Sound{"A", []Modifier{Lowered, SyllableDivision}}},
		{"95", Sound{Base: "95"}},
	} {
		t.Run(tc.source, func(t *testing.T) {
			doc := Parse(tc.source)
			if len(doc.Diagnostics) != 0 {
				t.Fatal(doc.Diagnostics)
			}
			if len(doc.Tokens) != 1 || !reflect.DeepEqual(doc.Tokens[0].Value, tc.want) {
				t.Fatalf("got %#v, want %#v", doc.Tokens, tc.want)
			}
		})
	}
}

func TestRenderSemanticModifiers(t *testing.T) {
	for _, mods := range [][]Modifier{{SemiHushed, Palatalized}, {Palatalized, SemiHushed}} {
		doc := Document{Tokens: []Token{{Value: Sound{"C", mods}}}}
		if got := Render(doc); got != "tʂ̻ʲ" {
			t.Fatal(got)
		}
	}
}

func TestParserPolicies(t *testing.T) {
	for _, tc := range []struct {
		source, want string
		diagnostics  int
	}{
		{"A47+,,", "ˈã̞̠", 0},
		{"A+74,,", "ˈã̞̠", 0},
		{"S87 S78", "ʂ̻ʲ ʂ̻ʲ", 0},
		{"A.,,", "ˈaː", 0},
		{"B95 A95 95", "bʔ a. ʔ", 0},
		{"(/041.020", " relevant to another question number041.020", 0},
		{"( 041.020 A", " relevant to problem number in dialectology 041.020 a", 0},
		{"?+", "? yes", 1},
		{"A++ A45 L5 B2- A*", "A++ A45 L5 B2- A*", 5},
		{"QUNKNOWN A94", "QUNKNOWN ă", 1},
		{"QEDNOPE A94", "QEDNOPE ă", 1},
		{"QENG Unicode: ייִדיש, C+ QADJ QP S78", "explanation in english:  Unicode: ייִדיש, C+ QADJ  ʂ̻ʲ", 0},
		{"QENG body without terminator", "explanation in english:  body without terminator", 0},
		{"A/B", "a/b", 0},
		{")", " prompted", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			doc := Parse(tc.source)
			if got := Render(doc); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if len(doc.Diagnostics) != tc.diagnostics {
				t.Fatalf("diagnostics: %#v", doc.Diagnostics)
			}
			assertCoverage(t, doc)
		})
	}
}

func assertCoverage(t *testing.T, doc Document) {
	t.Helper()
	pos := 0
	for _, token := range doc.Tokens {
		if token.Span.Start != pos || token.Span.End <= pos || token.Span.End > len(doc.Source) {
			t.Fatalf("gap, overlap, or invalid span: %#v", token)
		}
		pos = token.Span.End
	}
	if pos != len(doc.Source) {
		t.Fatalf("covered %d of %d bytes", pos, len(doc.Source))
	}
	for _, diagnostic := range doc.Diagnostics {
		if diagnostic.Span.Start < 0 || diagnostic.Span.End > len(doc.Source) || diagnostic.Span.Start >= diagnostic.Span.End {
			t.Fatal(diagnostic)
		}
	}
}

func FuzzSourceCoverage(f *testing.F) {
	for _, seed := range []string{"S78 A+,,", "QENG ייִדישQP A95", "QEDNOPE", "A++", "(/041.020", "", "//ABC"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		if !utf8.ValidString(source) {
			t.Skip()
		}
		doc := Parse(source)
		assertCoverage(t, doc)
		if !utf8.ValidString(Render(doc)) {
			t.Fatal("invalid rendered UTF-8")
		}
		if !reflect.DeepEqual(doc, Parse(source)) {
			t.Fatal("nondeterministic parse")
		}
	})
}

// This fixture belongs to the current converter; sharing it makes the draft's
// compatibility check independent of its production notation table.
func TestNotationCompatibility(t *testing.T) {
	fixture, err := os.ReadFile("../testdata/notation.tsv")
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(strings.TrimSuffix(string(fixture), "\n"), "\n") {
		source, want, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatalf("invalid fixture row %q", line)
		}
		t.Run(source, func(t *testing.T) {
			doc := Parse(source)
			if got := Render(doc); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
			if len(doc.Diagnostics) != 0 {
				t.Fatal(doc.Diagnostics)
			}
			assertCoverage(t, doc)
		})
	}
}
