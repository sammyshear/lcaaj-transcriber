package structured

import (
	"strings"
	"unicode/utf8"
)

var bases = map[string]string{
	"A": "a", "E": "e", "I": "i", "O": "o", "U": "u", "3": "ə", "1": "ɪ", "6": "ʌ",
	"B": "b", "C": "ts", "D": "d", "F": "f", "G": "g", "H": "h", "J": "j", "K": "k",
	"L": "l", "M": "m", "N": "n", "P": "p", "R": "r", "S": "s", "T": "t", "V": "v", "W": "w", "X": "x", "Z": "z",
	"95": "ʔ",
}

func in(base, class string) bool { return len(base) == 1 && strings.Contains(class, base) }
func vowel(base string) bool     { return in(base, "AEIOU316") }

func modifierFor(base, suffix string) (Modifier, bool) {
	switch suffix {
	case ".":
		return Long, true
	case "..":
		return ExtraLong, true
	case ",,":
		return PrimaryStress, vowel(base)
	case ",":
		if vowel(base) {
			return SecondaryStress, true
		}
		return Syllabic, true
	case "94":
		return Short, vowel(base)
	case "95":
		return SyllableDivision, vowel(base)
	case "4":
		return Lowered, vowel(base)
	case "5":
		return Raised, vowel(base)
	case "+":
		if vowel(base) {
			return Nasalized, true
		}
		if in(base, "CSZ") {
			return Hushed, true
		}
		return NasalRelease, in(base, "BDFGKPTV")
	case "2":
		if in(base, "BDGJLMNRVWZ") {
			return Voiceless, true
		}
		return Voiced, in(base, "CFHKPSTX")
	case "7":
		if vowel(base) {
			return Backed, true
		}
		if in(base, "CSZ") {
			return SemiHushed, true
		}
		return Velarized, in(base, "BDGJLMNRVWFHKPTX")
	case "8":
		if vowel(base) {
			return Fronted, true
		}
		return Palatalized, in(base, "BDGJLMNRVWFHKPTXSCZ")
	}
	return "", false
}

// Grouping makes conflicts explicit instead of resolving them by replacement
// order. These policies can be reviewed without changing the renderer.
func modifierGroup(m Modifier) Modifier {
	switch m {
	case Long, ExtraLong, Short:
		return "duration"
	case PrimaryStress, SecondaryStress:
		return "stress"
	case Hushed, SemiHushed:
		return "sibilant-quality"
	case Backed, Fronted:
		return "vowel-position"
	case Lowered, Raised:
		return "vowel-height"
	case Voiced, Voiceless:
		return "voicing"
	}
	return m
}

func parseSound(source string, pos int) (end int, value Node, problem string, ok bool) {
	start := pos
	r, size := utf8.DecodeRuneInString(source[pos:])
	base := strings.ToUpper(string(r))
	if strings.HasPrefix(source[pos:], "95") {
		base, size = "95", 2
	}
	if _, found := bases[base]; !found {
		return pos, nil, "", false
	}
	pos += size
	sound := Sound{Base: base}
	seen := map[Modifier]bool{}
	for pos < len(source) {
		suffix := ""
		// Unsupported but recognizable suffixes (2-, L5, *) are consumed as a
		// whole sound and preserved with a diagnostic, rather than partly converted.
		for _, candidate := range []string{"94", "95", "2-", ",,", "..", "+", "2", "4", "5", "7", "8", ",", ".", "*"} {
			if strings.HasPrefix(source[pos:], candidate) {
				suffix = candidate
				break
			}
		}
		if suffix == "" {
			break
		}
		mod, valid := modifierFor(base, suffix)
		// After consonants, 95 starts the next sound; after vowels, it is a suffix.
		if suffix == "95" && !valid {
			break
		}
		if !valid {
			problem = "unsupported modifier " + suffix + " on " + base
		} else {
			group := modifierGroup(mod)
			if seen[group] {
				problem = "duplicate or conflicting " + string(group) + " modifiers"
			}
			seen[group] = true
			sound.Modifiers = append(sound.Modifiers, mod)
		}
		pos += len(suffix)
		if mod == SyllableDivision {
			break
		}
	}
	if problem != "" {
		return pos, Unknown{source[start:pos]}, problem, true
	}
	return pos, sound, "", true
}
