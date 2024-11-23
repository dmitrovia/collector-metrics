package notallowedhandler

import (
	"fmt"
	"net/http"
)

type NotAllowedHandler struct{}

func (h NotAllowedHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	MethodNotAllowedHandler(rw, r)
}

func MethodNotAllowedHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusNotFound)

	Body := "Method not allowed!\n"
	fmt.Fprintf(rw, "%s", Body)
}
