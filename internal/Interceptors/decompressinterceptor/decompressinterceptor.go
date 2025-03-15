package decompressinterceptor

import (
	"bytes"
	"context"

	"github.com/dmitrovia/collector-metrics/internal/functions/compress"
	pb "github.com/dmitrovia/collector-metrics/pkg/microservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DecompressInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		reqType, ok := req.(*pb.SenderRequest)
		if !ok {
			return nil, status.Errorf(codes.Unknown,
				"DecryptInterceptor->Cast")
		}

		r := bytes.NewReader(reqType.GetMetrics())

		decompress, err := compress.DeflateDecompress(r)
		if err != nil {
			return nil, status.Errorf(codes.Unknown,
				"DecryptInterceptor->DeflateDecompress")
		}

		reqType.Metrics = decompress

		return handler(ctx, req)
	}
}
