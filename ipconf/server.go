package ipconf

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/ipconf/domain"
	"github.com/kuan525/tiger/ipconf/source"
)

func RunMain(path string) {
	config.Init(path)
	source.Init() // 数据源要先初始化
	domain.Init() // 初始化调度层
	s := server.Default(server.WithHostPorts(":6789"))
	s.GET("/ip/list", GetIpInfoList)
	s.Spin()
}
