package grpc

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func (s *GRPCServer) accessTokenInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	if info.FullMethod == "/gophkeeper.service.GophKeeperService/AddEntry" {

		var accessToken string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("access_token")
			if len(values) > 0 {
				accessToken = values[0]
			}
		}
		if len(accessToken) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		userId, err := auth.GetUserIDFromToken(accessToken, s.jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "err")
		}

		ctx = context.WithValue(ctx, userIDKey, userId)

	}

	return handler(ctx, req)
}
