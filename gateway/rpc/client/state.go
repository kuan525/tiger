package client

import (
	"context"
	"fmt"
	"time"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/tgrpc"
	"github.com/kuan525/tiger/state/rpc/service"
)

var stateClient service.StateClient

func initStateClient() {
	pCli, err := tgrpc.NewClient(config.GetStateServiceName())
	if err != nil {
		panic(err)
	}
	cli, err := pCli.DialByEndPoint(config.GetGatewayStateServerEndPoint())
	if err != nil {
		panic(err)
	}
	stateClient = service.NewStateClient(cli)
}

func CancelConn(ctx *context.Context, endpoint string, connID uint64, playLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	stateClient.CancelConn(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     playLoad,
	})
	return nil
}

func SendMsg(ctx *context.Context, endpoint string, connID uint64, playLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	fmt.Println("sendMsg", connID, string(playLoad))
	_, err := stateClient.SendMsg(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     playLoad,
	})
	if err != nil {
		panic(err)
	}
	return nil
}
