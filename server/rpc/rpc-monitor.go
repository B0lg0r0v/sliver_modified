package rpc

import (
	"context"

	"github.com/B0lg0r0v/sliver_modified/protobuf/commonpb"
	"github.com/B0lg0r0v/sliver_modified/server/configs"
	"github.com/B0lg0r0v/sliver_modified/server/watchtower"
)

func (rpc *Server) MonitorStart(ctx context.Context, _ *commonpb.Empty) (*commonpb.Response, error) {
	resp := &commonpb.Response{}
	config := configs.GetServerConfig()
	err := watchtower.StartWatchTower(config)
	if err != nil {
		resp.Err = err.Error()
	}
	return resp, err
}

func (rpc *Server) MonitorStop(ctx context.Context, _ *commonpb.Empty) (*commonpb.Empty, error) {
	resp := &commonpb.Empty{}
	watchtower.StopWatchTower()
	return resp, nil
}
