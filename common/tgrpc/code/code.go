package code

import "google.golang.org/grpc/codes"

const (
	CodeTooManyRequest codes.Code = 100
	CodeCircuitBreak   codes.Code = 101
)
