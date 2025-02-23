// Package setmetrichandler provides handler
// to check the functionality of the database.
package pinghandler

import (
	"context"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/jackc/pgx/v5"
)

// PingHandler - describing the handler.
type PingHandler struct {
	serv   service.Service
	params *bizmodels.InitParams
}

// NewPingHandler - to create an instance
// of a handler object.
func NewPingHandler(
	serv service.Service, par *bizmodels.InitParams,
) *PingHandler {
	return &PingHandler{serv: serv, params: par}
}

// PingHandler - main handler method.
func (h *PingHandler) PingHandler(
	writer http.ResponseWriter, req *http.Request,
) {
	ctx, cancel := context.WithTimeout(
		req.Context(), h.params.WaitSecRespDB)

	defer cancel()

	conn, err := pgx.Connect(ctx, h.params.DatabaseDSN)
	if err != nil {
		// fmt.Fprintf(os.Stderr,
		// "Unable to connect to database: %w\n", err)
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	defer conn.Close(ctx)
	writer.WriteHeader(http.StatusOK)
}
