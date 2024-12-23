package rpc

import (
	"context"

	"github.com/B0lg0r0v/sliver_modified/protobuf/commonpb"
	"github.com/B0lg0r0v/sliver_modified/protobuf/sliverpb"
)

// StartService creates and starts a Windows service on a remote host
func (rpc *Server) StartService(ctx context.Context, req *sliverpb.StartServiceReq) (*sliverpb.ServiceInfo, error) {
	resp := &sliverpb.ServiceInfo{Response: &commonpb.Response{}}
	err := rpc.GenericHandler(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// StopService stops a remote service
func (rpc *Server) StopService(ctx context.Context, req *sliverpb.StopServiceReq) (*sliverpb.ServiceInfo, error) {
	resp := &sliverpb.ServiceInfo{Response: &commonpb.Response{}}
	err := rpc.GenericHandler(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RemoveService deletes a service from the remote system
func (rpc *Server) RemoveService(ctx context.Context, req *sliverpb.RemoveServiceReq) (*sliverpb.ServiceInfo, error) {
	resp := &sliverpb.ServiceInfo{Response: &commonpb.Response{}}
	err := rpc.GenericHandler(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
