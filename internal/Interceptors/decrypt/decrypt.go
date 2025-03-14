package decrypt

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
			return nil, status.Errorf(codes.Unknown, "Ошибка")
		}

		reqType, ok := req.(*pb.SenderResponse)
		if !ok {
			return nil, status.Errorf(codes.Unknown, "Ошибка")
		}

		decr, err := asymcrypto.Decrypt(&reqType.Metrics, &key)
		if err == nil {
			reqType.Metrics = *decr
		}

		return handler(ctx, req)
	}
}
