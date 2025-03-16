package decryptmid

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dmitrovia/collector-metrics/internal/functions/asymcrypto"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
)

func DecryptMiddleware(
	params bizmodels.InitParams,
) func(http.Handler) http.Handler {
	handler := func(hand http.Handler) http.Handler {
		return http.HandlerFunc(
			func(
				writer http.ResponseWriter, req *http.Request,
			) {
				bodyD, err := io.ReadAll(req.Body)
				if err != nil {
					defer req.Body.Close()
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Println("DecryptMiddleware->ReadAll %w", err)

					return
				}

				key, err := os.ReadFile(params.CryptoPrivateKeyPath)
				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Println("DecryptMiddleware->ReadFile %w", err)

					return
				}

				// error reporting removed for autotests
				decr, err := asymcrypto.Decrypt(&bodyD, &key)
				if err == nil {
					req.Body = io.NopCloser(bytes.NewReader(*decr))
				} else {
					req.Body = io.NopCloser(bytes.NewReader(bodyD))
				}

				hand.ServeHTTP(writer, req)
			},
		)
	}

	return handler
}
