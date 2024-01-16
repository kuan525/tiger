package ipconf

import "github.com/kuan525/tiger/ipconf/domain"

func top5Endports(eds []*domain.Endpoint) []*domain.Endpoint {
	if len(eds) < 5 {
		return eds
	}
	return eds[:5]
}

func packRes(ed []*domain.Endpoint) Response {
	return Response{
		Message: "ok",
		Code:    0,
		Data:    ed,
	}
}
