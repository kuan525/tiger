package client

import (
	"context"
	"fmt"
	"time"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/tgrpc"
	"github.com/kuan525/tiger/gateway/rpc/service"
)

var gatewayClient service.GatewayClient

func initGatewayClient() {
	pCli, err := tgrpc.NewClient(config.GetGatewayServiceName())
	if err != nil {
		panic(err)
	}
	conn, err := pCli.DialByEndPoint(config.GetStateServerGatewayServerEndpoint())
	if err != nil {
		panic(err)
	}
	gatewayClient = service.NewGatewayClient(conn)
}

func DelConn(ctx *context.Context, connID uint64, payLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	gatewayClient.DelConn(rpcCtx, &service.GatewayRequest{ConnID: connID, Data: payLoad})
	return nil
}

func Push(ctx *context.Context, connID uint64, payLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	resp, err := gatewayClient.Push(rpcCtx, &service.GatewayRequest{ConnID: connID, Data: payLoad})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
	return nil
}
