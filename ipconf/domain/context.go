package domain

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type ClientContext struct {
	IP string `json:"ip"`
}

type IpConfContext struct {
	Ctx       *context.Context
	AppCtx    *app.RequestContext
	ClientCtx *ClientContext
}

func BuildIpConfContext(c *context.Context, ctx *app.RequestContext) *IpConfContext {
	return &IpConfContext{
		Ctx:       c,
		AppCtx:    ctx,
		ClientCtx: &ClientContext{},
	}
}
