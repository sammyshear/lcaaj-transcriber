package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sammyshear/lcaaj-transcriber/views"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type dataSignal struct {
	Data string `json:"data"`
}

func Transcribe(w http.ResponseWriter, r *http.Request) {
	data := &dataSignal{}
	err := datastar.ReadSignals(r, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sb strings.Builder
	sse := datastar.NewSSE(w, r)

	fmt.Printf("Data: %s\n", data.Data)

	for i, char := range data.Data {
		switch char {
		case 'A':
			sb.WriteString("a")
		case 'E':
			sb.WriteString("e")
		case 'I':
			sb.WriteString("i")
		case 'O':
			sb.WriteString("o")
		case 'U':
			sb.WriteString("u")
		case '3':
			sb.WriteString("ә")
		case '1':
			sb.WriteString("ɽ")
		case '6':
			sb.WriteString("ʌ")
		case '.':
			sb.WriteString("ː")
		case '9':
			switch data.Data[i+1] {
			case '4':
				sb.WriteString("◌̆")
			case '5':
				sb.WriteString("ʔ")
			}
		case '4':
			continue
		case '5':
			continue
		case ',':
			sb.WriteString("ˌ")
		case '+':
			sb.WriteString("◌̃")
		case 'B':
			sb.WriteString("b")
		case 'C':
			sb.WriteString("c")
		case 'D':
			sb.WriteString("d")
		case 'F':
			sb.WriteString("f")
		case 'G':
			sb.WriteString("g")
		case 'H':
			sb.WriteString("h")
		case 'J':
			sb.WriteString("j")
		case 'K':
			sb.WriteString("k")
		case 'L':
			if data.Data[i+1] == '5' {
				sb.WriteString("ɬ")
			} else {
				sb.WriteString("l")
			}
		case 'M':
			sb.WriteString("m")
		case 'N':
			sb.WriteString("n")
		case 'P':
			sb.WriteString("p")
		case 'R':
			sb.WriteString("r")
		case 'S':
			sb.WriteString("s")
		case 'T':
			sb.WriteString("t")
		case 'V':
			sb.WriteString("v")
		case 'W':
			sb.WriteString("w")
		case 'X':
			sb.WriteString("x")
		case 'Z':
			sb.WriteString("z")
		default:
			sb.WriteRune(char)
		}
	}

	output := sb.String()

	fmt.Printf("Result: %s\n", output)

	sse.MergeFragmentTempl(views.Transcription(output), datastar.WithSelectorID("result"))
}
