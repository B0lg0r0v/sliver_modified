package rpc

import (
	"context"

	"github.com/B0lg0r0v/sliver_modified/protobuf/commonpb"
	"github.com/B0lg0r0v/sliver_modified/protobuf/sliverpb"
)

// Ping - Try to send a round trip message to the implant
func (rpc *Server) Ping(ctx context.Context, req *sliverpb.Ping) (*sliverpb.Ping, error) {
	resp := &sliverpb.Ping{Response: &commonpb.Response{}}
	err := rpc.GenericHandler(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
