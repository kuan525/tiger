package state

import (
	"context"
	"fmt"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/tgrpc"
	"github.com/kuan525/tiger/state/rpc/client"
	"github.com/kuan525/tiger/state/rpc/service"
	"google.golang.org/grpc"
)

var cmdChannel chan *service.CmdContext

func RunMain(path string) {
	config.Init(path)
	cmdChannel = make(chan *service.CmdContext, config.GetStateCmdChannelNum())

	s := tgrpc.NewServer(
		tgrpc.WithServiceName(config.GetStateServiceName()),
		tgrpc.WithIP(config.GetStateServiceAddr()),
		tgrpc.WithPort(config.GetStateServicePort()),
		tgrpc.WithWeight(config.GetStateRPCWeight()))

	s.RegisterService(func(server *grpc.Server) {
		service.RegisterStateServer(server, &service.Service{CmdChannel: cmdChannel})
	})

	client.Init()
	go cmdHandler()
	s.Start(context.TODO())
}

func cmdHandler() {
	for cmd := range cmdChannel {
		switch cmd.Cmd {
		case service.CancelConnCmd:
			fmt.Printf("cancelconn endpoint:%s, fd:%d, data:%+v", cmd.Endpoint, cmd.FD, cmd.PlayLoad)
		case service.SendMsgCmd:
			fmt.Println("cmdHandler", string(cmd.PlayLoad))
			client.Push(cmd.Ctx, int32(cmd.FD), cmd.PlayLoad)
		}
	}
}
