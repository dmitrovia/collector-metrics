package checkipmid

import (
	"fmt"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/ip"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

func CheckIPMiddleware(
	params bizmodels.InitParams,
) func(http.Handler) http.Handler {
	handler := func(hand http.Handler) http.Handler {
		return http.HandlerFunc(
			func(
				writer http.ResponseWriter, req *http.Request,
			) {
				realIP := req.Header.Get("X-Real-IP")

				if realIP != "" && params.TrustedSubnet != "" {
					isC, err := ip.ContainsIPInSubnet(realIP,
						params.TrustedSubnet)
					if err != nil {
						writer.WriteHeader(http.StatusInternalServerError)
						fmt.Println(" %w",
							err)

						return
					}

					if !isC {
						writer.WriteHeader(http.StatusForbidden)

						return
					}
				}

				hand.ServeHTTP(writer, req)
			},
		)
	}

	return handler
}
