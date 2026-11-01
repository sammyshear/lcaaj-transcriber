package structured

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Parse reads original source only. Each byte belongs to exactly one token,
// including whitespace, controls, and unsupported notation. Comment bodies are
// consumed in literal mode and never inspected for sounds or editorial labels.
func Parse(source string) Document {
	doc := Document{Source: source}
	emit := func(start, end int, value Node) { doc.Tokens = append(doc.Tokens, Token{Span{start, end}, value}) }
	for pos := 0; pos < len(source); {
		start := pos
		if rule, end, ok := matchNote(source, pos); ok {
			pos = end
			if rule.code == "QP" {
				emit(start, pos, Control{rule.code})
			} else {
				emit(start, pos, Note{rule.code, rule.label})
			}
			if rule.body == literalBody {
				end = pos
				if offset := strings.Index(source[pos:], "QP"); offset >= 0 {
					end = pos + offset
				} else {
					end = len(source)
				}
				if end > pos {
					emit(pos, end, Literal{source[pos:end]})
				}
				pos = end
				// Consume the terminator directly, so a literal body's final character
				// cannot affect lexical decisions about returning to phonetic mode.
				if pos < len(source) {
					emit(pos, pos+2, Control{"QP"})
					pos += 2
				}
			} else if isReference(rule.code) {
				// Reference numbers are not phonetic digit vowels or editorial zeroes.
				end = referenceEnd(source, pos)
				if end > pos {
					emit(pos, end, Literal{source[pos:end]})
					pos = end
				}
			}
			continue
		}
		if strings.HasPrefix(source[pos:], "//") {
			pos += 2
			body := pos
			for pos < len(source) && asciiWord(source[pos]) {
				pos++
			}
			emit(start, pos, Optional{source[body:pos]})
			continue
		}
		if end, value, problem, ok := parseSound(source, pos); ok {
			emit(pos, end, value)
			if problem != "" {
				doc.Diagnostics = append(doc.Diagnostics, Diagnostic{Span{pos, end}, problem})
			}
			pos = end
			continue
		}
		r, size := utf8.DecodeRuneInString(source[pos:])
		pos += size
		if unicode.IsSpace(r) || r == '/' || r == ';' || r == ':' {
			emit(start, pos, Literal{source[start:pos]})
			continue
		}
		if r == 'Q' {
			// Preserve an unknown Q-code as a unit instead of transcribing its letters.
			for pos < len(source) {
				next, n := utf8.DecodeRuneInString(source[pos:])
				if unicode.IsSpace(next) || next == '/' {
					break
				}
				pos += n
			}
		}
		emit(start, pos, Unknown{source[start:pos]})
		doc.Diagnostics = append(doc.Diagnostics, Diagnostic{Span{start, pos}, "unsupported source notation"})
	}
	return doc
}

// The longest marker wins at the current position. Unlike the regex converter,
// the parser never consumes arbitrary punctuation as a marker's left boundary.
// One leading whitespace rune is absorbed for legacy labels that already carry
// a leading space; this retains existing output spacing without losing source.
func matchNote(source string, pos int) (noteRule, int, bool) {
	var best noteRule
	end := pos
	for _, rule := range noteRules {
		offset := pos
		if rule.leadingBoundary {
			r, size := utf8.DecodeRuneInString(source[pos:])
			if unicode.IsSpace(r) {
				offset += size
			} else if pos > 0 {
				previous, _ := utf8.DecodeLastRuneInString(source[:pos])
				if previous == ')' || (previous < 128 && asciiWord(byte(previous))) {
					continue
				}
			}
		}
		if !strings.HasPrefix(source[offset:], rule.code) {
			continue
		}
		candidateEnd := offset + len(rule.code)
		if rule.code == "0" {
			if (pos > 0 && asciiWord(source[pos-1])) || (candidateEnd < len(source) && asciiWord(source[candidateEnd])) {
				continue
			}
		}
		// Reject an unknown Q-code instead of accepting an arbitrary known prefix.
		// QP remains valid without delimiters,
		// matching the current converter's termination convention.
		if strings.HasPrefix(rule.code, "Q") && rule.code != "QP" &&
			candidateEnd < len(source) && asciiLetter(source[candidateEnd]) {
			continue
		}
		if len(rule.code) > len(best.code) {
			best, end = rule, candidateEnd
		}
	}
	return best, end, best.code != ""
}

func asciiLetter(b byte) bool { return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' }
func asciiWord(b byte) bool   { return asciiLetter(b) || b >= '0' && b <= '9' }
func isReference(code string) bool {
	return code == "(" || code == "(/" || code == "($" || code == "(("
}

func referenceEnd(source string, pos int) int {
	start := pos
	for pos < len(source) {
		r, n := utf8.DecodeRuneInString(source[pos:])
		if !unicode.IsSpace(r) {
			break
		}
		pos += n
	}
	digits := pos
	for pos < len(source) && (source[pos] >= '0' && source[pos] <= '9' || source[pos] == '.') {
		pos++
	}
	if pos == digits || source[digits] == '.' {
		return start
	}
	return pos
}
