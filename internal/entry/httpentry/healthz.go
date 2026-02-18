package httpentry

import (
	"net/http"
)

var (
	aliveResp = []byte(`{"status":"im ok"}`)
	readyResp = []byte(`{"status":"im ready"}`)
)

func aliveHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(aliveResp)
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(readyResp)
}
