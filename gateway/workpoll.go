package gateway

import (
	"fmt"

	"github.com/kuan525/tiger/common/config"
	"github.com/panjf2000/ants"
)

var wPool *ants.Pool

// 初始化协程池
func initWorkPool() {
	var err error
	if wPool, err = ants.NewPool(config.GetGatewayWorkerPoolNum()); err != nil {
		fmt.Printf("initWorkPool.err :%s num:%d\n", err.Error(), config.GetGatewayWorkerPoolNum())
	}
}
