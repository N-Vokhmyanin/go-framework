package ctxlog

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

const userAgentField = "user-agent"

func GetUserAgentFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		return getUserAgentFromMetadata(md)
	}
	return ""
}

func getUserAgentFromMetadata(md metadata.MD) string {
	var userAgents []string

	// user-agent from http request
	userAgents = md.Get(runtime.MetadataPrefix + userAgentField)
	if len(userAgents) > 0 {
		return userAgents[0]
	}

	// user-agent from grpc request
	userAgents = md.Get(userAgentField)
	if len(userAgents) > 0 {
		return userAgents[0]
	}

	return ""
}
