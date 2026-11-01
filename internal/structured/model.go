// Package structured is an experimental, regex-free LCAAJ converter. Parse
// retains source spans and diagnostics; Render converts parsed meaning to text.
// The application selects it only when built with -tags structured.
package structured

// Span is a half-open range of UTF-8 byte offsets in Document.Source.
type Span struct{ Start, End int }

type Token struct {
	Span  Span
	Value Node
}

// Node is a closed set of token values. Consumers can inspect concrete types
// without having to interpret rendered IPA or generated prose.
type Node interface{ isNode() }

type Sound struct {
	Base      string     // Canonical source symbol, e.g. A, C, 3, or 95.
	Modifiers []Modifier // Semantic modifiers, retained in source order.
}
type (
	Note     struct{ Code, Label string }
	Literal  struct{ Text string }
	Control  struct{ Code string }
	Optional struct{ Text string } // Existing //TEXT convention.
	Unknown  struct{ Text string }
)

func (Sound) isNode()    {}
func (Note) isNode()     {}
func (Literal) isNode()  {}
func (Control) isNode()  {}
func (Optional) isNode() {}
func (Unknown) isNode()  {}

type Diagnostic struct {
	Span    Span
	Message string
}
type Document struct {
	Source      string
	Tokens      []Token
	Diagnostics []Diagnostic
}

type Modifier string

const (
	Nasalized        Modifier = "nasalized"
	Hushed           Modifier = "hushed"
	SemiHushed       Modifier = "semi-hushed"
	NasalRelease     Modifier = "nasal-release"
	Palatalized      Modifier = "palatalized"
	Backed           Modifier = "backed"
	Fronted          Modifier = "fronted"
	Lowered          Modifier = "lowered"
	Raised           Modifier = "raised"
	Velarized        Modifier = "velarized"
	Voiced           Modifier = "voiced"
	Voiceless        Modifier = "voiceless"
	Short            Modifier = "short"
	Long             Modifier = "long"
	ExtraLong        Modifier = "extra-long"
	SecondaryStress  Modifier = "secondary-stress"
	PrimaryStress    Modifier = "primary-stress"
	Syllabic         Modifier = "syllabic"
	SyllableDivision Modifier = "syllable-division"
)
