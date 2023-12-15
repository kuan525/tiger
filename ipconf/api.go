package ipconf

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/kuan525/tiger/ipconf/domain"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

func GetIpInfoList(c context.Context, ctx *app.RequestContext) {
	defer func() {
		if err := recover(); err != nil {
			ctx.JSON(consts.StatusBadRequest, utils.H{"err": err})
		}
	}()
	// step1 构件客户请求信息
	ipConfCtx := domain.BuildIpConfContext(&c, ctx)
	// step2 进行ip调度
	eds := domain.Dispatch(ipConfCtx)
	// step3 根据得分取top5返回
	ipConfCtx.AppCtx.JSON(consts.StatusOK, packRes(top5Endports(eds)))
}
