package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	userAgentHeader            = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) ExtractMetadata(ctx context.Context) *Metadata {
	meta := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// md is mapping (string => []string)
		// log.Printf("md: %+v\n", md)

		// for requests coming from http clients
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}
		// for requests coming from grpc-cli clients like evans
		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}
		// for requests coming from http clients
		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			meta.ClientIP = clientIPs[0]
		}
	}
	// for requests coming from grpc-cli clients like evans
	if peer, ok := peer.FromContext(ctx); ok {
		// log.Printf("peer: %+v\n", peer)
		meta.ClientIP = peer.Addr.String()
	}
	return meta
}
