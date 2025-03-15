package decryptinterceptor

import (
	"context"
	"os"

	"github.com/dmitrovia/collector-metrics/internal/functions/asymcrypto"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	pb "github.com/dmitrovia/collector-metrics/pkg/microservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const cun = codes.Unknown

func DecryptInterceptor(
	params *bizmodels.InitParams,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		key, err := os.ReadFile(params.CryptoPrivateKeyPath)
		if err != nil {
			return nil, status.Errorf(cun, "DecryptInterceptor->RF")
		}

		reqType, ok := req.(*pb.SenderRequest)
		if !ok {
			return nil, status.Errorf(cun, "DecryptInterceptor->C")
		}

		decr, err := asymcrypto.Decrypt(&reqType.Metrics, &key)
		if err == nil {
			reqType.Metrics = *decr
		}

		return handler(ctx, req)
	}
}
