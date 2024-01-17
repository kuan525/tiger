package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/tcp"
	"github.com/kuan525/tiger/common/tgrpc"
	"github.com/kuan525/tiger/gateway/rpc/client"
	"github.com/kuan525/tiger/gateway/rpc/service"
	"google.golang.org/grpc"
)

// 接受state指令的通道
var cmdChannel chan *service.CmdContext

// 启动网关服务 [配置文件路径]
func RunMain(path string) {
	config.Init(path)
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: config.GetGatewayTCPServerPort()})
	if err != nil {
		log.Fatalf("StartTCPEPollServer err:%s", err.Error())
		panic(err)
	}
	initWorkPool()
	initEpoll(ln, runProc)
	fmt.Println("-------------im gateway stated-------------")

	cmdChannel = make(chan *service.CmdContext, config.GetGatewayCmdChannelNum())
	s := tgrpc.NewServer(
		tgrpc.WithServiceName(config.GetGatewayServiceName()),
		tgrpc.WithIP(config.GetGatewayServiceAddr()),
		tgrpc.WithPort(config.GetGatewayRPCServerPort()),
		tgrpc.WithWeight(config.GetGatewayRPCWeight()),
	)
	fmt.Println(config.GetGatewayServiceName(), config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort(), config.GetGatewayRPCWeight())
	s.RegisterService(func(server *grpc.Server) {
		service.RegisterGatewayServer(server, &service.Service{CmdChannel: cmdChannel})
	})
	// 启动rpc客户端
	client.Init()
	go cmdHandler()
	// 启动rpc server
	s.Start(context.TODO())
}

func runProc(c *connection, eper *epoller) {
	ctx := context.Background()
	// step1: 读取一个完整的包
	dataBuf, err := tcp.ReadData(c.conn)
	if err != nil {
		// 如果读取conn时发现连接关闭，则直接关闭端口连接
		// 通知state清理掉意外退出的conn的状态信息
		if errors.Is(err, io.EOF) {
			// 这个操作是异步的，不需要等到返回成功再执行，因为消息可靠性的保障是通过协议完成的而非某次cmd
			eper.remove(c)
			client.CancelConn(&ctx, getEndpoint(), c.id, nil)
		}
		return
	}
	err = wPool.Submit(func() {
		// step2:交给state server rpc处理
		client.SendMsg(&ctx, getEndpoint(), c.id, dataBuf)
	})
	if err != nil {
		fmt.Errorf("runProc:err:%+v\n", err.Error())
	}
}

func cmdHandler() {
	for cmd := range cmdChannel {
		// 异步提交到协程池中完成发送任务
		switch cmd.Cmd {
		case service.DelConnCmd:
			wPool.Submit(func() { closeConn(cmd) })
		case service.PushCmd:
			wPool.Submit(func() { sendMsgByCmd(cmd) })
		default:
			panic("command undefined")
		}
	}
}

func closeConn(cmd *service.CmdContext) {
	if connPtr, ok := ep.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		conn.Close()
	}
}

func sendMsgByCmd(cmd *service.CmdContext) {
	if connPtr, ok := ep.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		dp := tcp.DataPkg{
			Len:  uint32(len(cmd.Payload)),
			Data: cmd.Payload,
		}
		tcp.SendData(conn.conn, dp.Marshal())
	}
}

func getEndpoint() string {
	return fmt.Sprintf("%s:%d", config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort())
}
