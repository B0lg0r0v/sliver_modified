package rpc

import (
	"github.com/B0lg0r0v/sliver_modified/protobuf/clientpb"
	"github.com/B0lg0r0v/sliver_modified/protobuf/commonpb"
	"github.com/B0lg0r0v/sliver_modified/protobuf/rpcpb"
	"github.com/B0lg0r0v/sliver_modified/server/core"
	"github.com/B0lg0r0v/sliver_modified/server/log"
)

var (
	rpcEventsLog = log.NamedLogger("rpc", "events")
)

// Events - Stream events to client
func (rpc *Server) Events(_ *commonpb.Empty, stream rpcpb.SliverRPC_EventsServer) error {
	commonName := rpc.getClientCommonName(stream.Context())
	client := core.NewClient(commonName)
	core.Clients.Add(client)
	events := core.EventBroker.Subscribe()

	defer func() {
		rpcEventsLog.Infof("%d client disconnected", client.ID)
		core.EventBroker.Unsubscribe(events)
		core.Clients.Remove(client.ID)
	}()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event := <-events:
			pbEvent := &clientpb.Event{
				EventType: event.EventType,
				Data:      event.Data,
			}

			if event.Job != nil {
				pbEvent.Job = event.Job.ToProtobuf()
			}
			if event.Client != nil {
				pbEvent.Client = event.Client.ToProtobuf()
			}
			if event.Session != nil {
				pbEvent.Session = event.Session.ToProtobuf()
			}
			if event.Err != nil {
				pbEvent.Err = event.Err.Error()
			}

			err := stream.Send(pbEvent)
			if err != nil {
				rpcEventsLog.Warnf(err.Error())
				return err
			}
		}
	}
}
