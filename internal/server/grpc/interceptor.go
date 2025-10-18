package grpc

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) accessTokenInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	if info.FullMethod == "/gophkeeper.service.GophKeeperService/Sync" {

		var accessToken string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get(common.AccessTokenHeaderName)
			if len(values) > 0 {
				accessToken = values[0]
			}
		}
		if len(accessToken) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		userId, err := auth.GetUserIDFromToken(accessToken, s.jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		ctx = context.WithValue(ctx, shared.UserIDKey, userId)

	}

	return handler(ctx, req)
}
