// Package notallowedhandler provides handler
// default output for unavailable methods.
package notallowedhandler

import (
	"fmt"
	"net/http"
)

// NotAllowedHandler - describing the handler.
type NotAllowedHandler struct{}

func (h NotAllowedHandler) ServeHTTP(
	rw http.ResponseWriter, r *http.Request,
) {
	MethodNotAllowedHandler(rw, r)
}

// MethodNotAllowedHandler - main handler method.
func MethodNotAllowedHandler(
	rw http.ResponseWriter, _ *http.Request,
) {
	rw.WriteHeader(http.StatusNotFound)

	Body := "Method not allowed!\n"
	fmt.Fprintf(rw, "%s", Body)
}
