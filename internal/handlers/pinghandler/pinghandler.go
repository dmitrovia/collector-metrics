package pinghandler

import (
	"context"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/jackc/pgx/v5"
)

type PingHandler struct {
	serv   service.Service
	params *bizmodels.InitParams
}

func NewPingHandler(serv service.Service, par *bizmodels.InitParams) *PingHandler {
	return &PingHandler{serv: serv, params: par}
}

func (h *PingHandler) PingHandler(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), h.params.WaitSecRespDB)

	defer cancel()

	conn, err := pgx.Connect(ctx, h.params.DatabaseDSN)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		writer.WriteHeader(http.StatusInternalServerError)

		defer conn.Close(ctx)

		return
	}

	defer conn.Close(ctx)
	writer.WriteHeader(http.StatusOK)
}
