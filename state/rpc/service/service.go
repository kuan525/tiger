package service

import "context"

const (
	CancelConnCmd = 1
	SendMsgCmd    = 2
)

type CmdContext struct {
	Ctx      *context.Context
	Cmd      int32
	Endpoint string
	FD       int
	PlayLoad []byte
}

type Service struct {
	CmdChannel chan *CmdContext
}

func (s *Service) CancelConn(ctx context.Context, sr *StateRequest) (*StateResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:      &c,
		Cmd:      CancelConnCmd,
		FD:       int(sr.GetFd()),
		Endpoint: sr.GetEndpoint(),
	}
	return &StateResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}

func (s *Service) SendMsg(ctx context.Context, sr *StateRequest) (*StateResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx:      &c,
		Cmd:      SendMsgCmd,
		FD:       int(sr.GetFd()),
		Endpoint: sr.GetEndpoint(),
		PlayLoad: sr.GetData(),
	}
	return &StateResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
