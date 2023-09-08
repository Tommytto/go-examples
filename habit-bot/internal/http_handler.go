package internal

import (
	"net/http"
)

type HttpHandler struct {
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

func (h *HttpHandler) Healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ok"))
	w.WriteHeader(http.StatusOK)
}

func (h *HttpHandler) HandleOptions(w http.ResponseWriter) {
	writeCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}
func writeCORSHeaders(w http.ResponseWriter) {
	// Force build
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
}
