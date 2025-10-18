// Package grpc contains the server's gRPC transport glue: interceptors,
// request context enrichment, and RPC handlers.
package grpc

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// accessTokenInterceptor is a unary server interceptor that enforces access-token
// authentication for selected methods and injects the authenticated user ID into
// the request context.
//
// Current behavior:
//   - Only the Sync RPC ("/gophkeeper.service.GophKeeperService/Sync") is protected.
//   - The interceptor looks for the access token in gRPC metadata under
//     common.AccessTokenHeaderName.
//   - On success, it parses the token, extracts the user ID, and stores it in
//     the context under UserIDKey, then calls the handler.
//   - On failure, it returns codes.Unauthenticated.
//
// Note: If you add more authenticated methods, extend the method check accordingly.
func (s *GRPCServer) accessTokenInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
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

		userID, err := auth.GetUserIDFromToken(accessToken, s.jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		ctx = context.WithValue(ctx, UserIDKey, userID)
	}
	return handler(ctx, req)
}
