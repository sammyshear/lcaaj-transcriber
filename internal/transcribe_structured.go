//go:build structured

package internal

import "github.com/sammyshear/lcaaj-transcriber/internal/structured"

// Transcribe opts into the experimental parser with -tags structured. Normal
// builds retain the regex converter. The existing HTTP contract returns only
// text; structured.Parse retains diagnostics and source positions internally.
func Transcribe(data *dataSignal) string {
	return structured.Transcribe(data.Data)
}
