//go:build structured

package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// These combinations deliberately exercise the experimental parser, rather than
// just checking that a handler returns text from either implementation.
const structuredRequest = `{"data":"S87 A47+,,"}`
const structuredResponse = "ʂ̻ʲ ˈã̞̠"

func TestStructuredRESTEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/transcribe", strings.NewReader(structuredRequest))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	APITranscribe(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if got := response.Body.String(); got != structuredResponse {
		t.Fatalf("got %q, want %q", got, structuredResponse)
	}
}

func TestStructuredFrontendEndpoint(t *testing.T) {
	query := url.Values{"datastar": {structuredRequest}}
	request := httptest.NewRequest(http.MethodGet, "/api/dtranscribe?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	DatastarTranscribe(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("unexpected content type %q", got)
	}
	if got := response.Body.String(); !strings.Contains(got, `<span id="result">`+structuredResponse+`</span>`) {
		t.Fatalf("missing frontend result in %q", got)
	}
}
