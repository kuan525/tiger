package code

import "google.golang.org/grpc/codes"

const (
	CodeTooManyRequest codes.Code = 100 // 请求过多，一个客户端在一定时间发送过多请求
	CodeCircuitBreak   codes.Code = 101 // 电路中断，网络错误
)
