package service

import "context"

const (
	DelConnCmd = 1 // DelConn
	PushCmd    = 2 // Push
)

type CmdContext struct {
	Ctx      *context.Context
	Cmd      int32
	FD       int
	Playload []byte
}

type Service struct {
	CmdChannel chan *CmdContext
}

func (s *Service) DelConn(ctx context.Context, gr *GatewayRequest) (*GatewayResponse, error) {
	c := context.TODO() // 防止上下文结束影响异步处理的协程
	s.CmdChannel <- &CmdContext{
		Ctx: &c,
		Cmd: DelConnCmd,
		FD:  int(gr.GetFd()),
	}
	return &GatewayResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}

func (s *Service) Push(ctx context.Context, gr *GatewayRequest) (*GatewayResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:      &c,
		Cmd:      PushCmd,
		FD:       int(gr.GetFd()),
		Playload: gr.GetData(),
	}
	return &GatewayResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
