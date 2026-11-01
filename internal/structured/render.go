package structured

import "strings"

// Transcribe is the convenience entry point for callers that only need text.
// Call Parse and Render separately to retain diagnostics and inspect tokens.
func Transcribe(source string) string { return Render(Parse(source)) }

func Render(doc Document) string {
	var out strings.Builder
	for _, token := range doc.Tokens {
		switch value := token.Value.(type) {
		case Sound:
			out.WriteString(renderSound(value))
		case Note:
			out.WriteString(value.Label)
		case Literal:
			out.WriteString(value.Text)
		case Optional:
			out.WriteString("(")
			out.WriteString(value.Text)
			out.WriteString(")")
		case Unknown:
			out.WriteString(value.Text)
		case Control: // QP changes parsing mode and produces no visible output.
		}
	}
	return out.String()
}

func renderSound(sound Sound) string {
	has := map[Modifier]bool{}
	for _, mod := range sound.Modifiers {
		has[mod] = true
	}
	base := bases[sound.Base]
	if has[Hushed] {
		base = map[string]string{"C": "tʃ", "S": "ʃ", "Z": "ʒ"}[sound.Base]
	}
	if has[SemiHushed] {
		base = map[string]string{"C": "tʂ̻", "S": "ʂ̻", "Z": "ʐ̻"}[sound.Base]
	}
	var out strings.Builder
	if has[PrimaryStress] {
		out.WriteString("ˈ")
	} else if has[SecondaryStress] {
		out.WriteString("ˌ")
	}
	out.WriteString(base)
	// Canonical order is independent of source modifier order. Complete affricates
	// remain one Sound, so marks cannot accidentally attach to only a regex match.
	for _, mark := range []struct {
		mod Modifier
		ipa string
	}{
		{Short, "̆"},
		{Nasalized, "̃"},
		{Lowered, "̞"},
		{Raised, "̝"},
		{Backed, "̠"},
		{Fronted, "̟"},
		{Voiceless, "̥"},
		{Voiced, "̬"},
		{Velarized, "ˠ"},
		{Palatalized, "ʲ"},
		{NasalRelease, "ⁿ"},
		{Syllabic, "̩"},
		{Long, "ː"},
		{ExtraLong, "ːː"},
		{SyllableDivision, "."},
	} {
		if has[mark.mod] {
			out.WriteString(mark.ipa)
		}
	}
	return out.String()
}
