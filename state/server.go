package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/idl/message"
	"github.com/kuan525/tiger/common/tgrpc"
	"github.com/kuan525/tiger/state/rpc/client"
	"github.com/kuan525/tiger/state/rpc/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func RunMain(path string) {
	config.Init(path)
	cmdChannel = make(chan *service.CmdContext, config.GetStateCmdChannelNum())
	connToStateTable = sync.Map{}

	s := tgrpc.NewServer(
		tgrpc.WithServiceName(config.GetStateServiceName()),
		tgrpc.WithIP(config.GetStateServiceAddr()),
		tgrpc.WithPort(config.GetStateServicePort()),
		tgrpc.WithWeight(config.GetStateRPCWeight()))

	s.RegisterService(func(server *grpc.Server) {
		service.RegisterStateServer(server, &service.Service{CmdChannel: cmdChannel})
	})

	client.Init()
	// 启动时间轮
	InitTimer()
	go cmdHandler()
	s.Start(context.TODO())
}

func cmdHandler() {
	for cmdCtx := range cmdChannel {
		switch cmdCtx.Cmd {
		case service.CancelConnCmd:
			fmt.Printf("cancelconn endpoint:%s, fd:%d, data:%+v", cmdCtx.Endpoint, cmdCtx.ConnID, cmdCtx.PayLoad)
		case service.SendMsgCmd:
			fmt.Println("cmdHandler", string(cmdCtx.PayLoad))
			msgCmd := &message.MsgCmd{}

			err := proto.Unmarshal(cmdCtx.PayLoad, msgCmd)
			if err != nil {
				fmt.Printf("SendMsgCmd:err=%s\n", err.Error())
			}
			msgCmdhandler(cmdCtx, msgCmd)
		}
	}
}

func msgCmdhandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	switch msgCmd.Type {
	case message.CmdType_Login:
		loginMsgHandler(cmdCtx, msgCmd)
	case message.CmdType_Heartbeat:
		hearbeatMsgHandler(cmdCtx, msgCmd)
	case message.CmdType_ReConn:
		reConnMsgHandler(cmdCtx, msgCmd)
	}
}

func reConnMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	reConnMsg := &message.ReConnMsg{}
	err := proto.Unmarshal(msgCmd.Payload, reConnMsg)
	if err != nil {
		fmt.Printf("reConnMsgHandler:err=%s\n", err.Error())
		return
	}

	// 重连的消息头中的connID才是上一次断开链接的connID
	if data, ok := connToStateTable.Load(reConnMsg.Head.ConnID); ok {
		state, _ := data.(*connState)
		state.Lock()
		defer state.Unlock()
		//先停止定时任务的回调
		if state.reConnTimer != nil {
			state.reConnTimer.Stop()
			state.reConnTimer = nil // 重连定时器被清除
		}
		// 从索引中删除旧的connID
		connToStateTable.Delete(reConnMsg.Head.ConnID)
		// 变更connID，cmdCTX中的connID才是gateway重连的新链接
		state.connID = cmdCtx.ConnID
		connToStateTable.Store(cmdCtx.ConnID, state)
		sendACKMsg(cmdCtx.ConnID, 0, "reconn ok")
	} else {
		sendACKMsg(cmdCtx.ConnID, 1, "reconn failed")
	}
}

func hearbeatMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	heartMsg := &message.HeartbeatMsg{}
	err := proto.Unmarshal(msgCmd.Payload, heartMsg)
	if err != nil {
		fmt.Printf("hearbeatMsgHandler:err=%s\n", err.Error())
		return
	}
	if data, ok := connToStateTable.Load(cmdCtx.ConnID); ok {
		sate, _ := data.(*connState)
		sate.reSetHeartTimer()
	}
	// 为减少通信量，可以暂时不回复心跳的ack
}

func loginMsgHandler(cmdCtx *service.CmdContext, msgCmd *message.MsgCmd) {
	loginMsg := &message.LoginMsg{}
	err := proto.Unmarshal(msgCmd.Payload, loginMsg)
	if err != nil {
		fmt.Printf("loginMsgHandler:err=%s\n", err.Error())
		return
	}
	if loginMsg.Head != nil {
		// 这里把login msg传送给业务层处理
		fmt.Println("loginMsgHandler", loginMsg.Head.DeviceID)
	}
	// 创建定时器
	t := AfterFunc(300*time.Second, func() {
		clearState(cmdCtx.ConnID)
	})
	// 初始化链接的状态
	connToStateTable.Store(cmdCtx.ConnID, &connState{heartTimer: t, connID: cmdCtx.ConnID})
}

func sendACKMsg(connID uint64, code uint32, msg string) {
	ackMsg := &message.ACKMsg{}
	ackMsg.Code = code
	ackMsg.Msg = msg
	ctx := context.TODO()
	downLoad, err := proto.Marshal(ackMsg)
	if err != nil {
		fmt.Println("sendACKMsg", err)
	}
	client.Push(&ctx, connID, downLoad)
}
