// Package internal provides the internal backend code
package internal

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sammyshear/lcaaj-transcriber/views"
	"github.com/starfederation/datastar-go/datastar"
)

type dataSignal struct {
	Data string `json:"data"`
}

func APITranscribe(w http.ResponseWriter, r *http.Request) {
	data := &dataSignal{}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = json.Unmarshal(b, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	_, _ = w.Write([]byte(Transcribe(data)))
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
	_ = sse.PatchElementTempl(views.Transcription(o), datastar.WithSelectorID("result"))
}
