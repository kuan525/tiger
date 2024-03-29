package source

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/discovery"
)

// 模拟注册节点，并且不断更新
func testServiceRegister(ctx *context.Context, port, node string) {
	// 模拟服务发现
	go func() {
		ed := discovery.EndpointInfo{
			IP:   "127.0.0.1",
			Port: port,
			MetaData: map[string]interface{}{
				"connect_num":   float64(rand.Int63n(12312321231231131)),
				"message_bytes": float64(rand.Int63n(1231232131556)),
			},
		}
		sr, err := discovery.NewServiceRegister(ctx, fmt.Sprintf("%s/%s", config.GetServicePathForIpConf(), node), &ed, time.Now().Unix())
		if err != nil {
			panic(err)
		}
		go sr.ListenLeaseRespChan()
		for {
			ed = discovery.EndpointInfo{
				IP:   "127.0.0.1",
				Port: port,
				MetaData: map[string]interface{}{
					"connect_num":   float64(rand.Int63n(12312321231231131)),
					"message_bytes": float64(rand.Int63n(1231232131556)),
				},
			}
			sr.UpdateValue(&ed)
			time.Sleep(1 * time.Second)
		}
	}()
}

func testServiceRegisterClose(ctx *context.Context, port, node string) {
	ed := discovery.EndpointInfo{
		IP:   "127.0.0.1",
		Port: port,
		MetaData: map[string]interface{}{
			"connect_num":   float64(rand.Int63n(12312321231231131)),
			"message_bytes": float64(rand.Int63n(1231232131556)),
		},
	}
	sr, err := discovery.NewServiceRegister(ctx, fmt.Sprintf("%s/%s", config.GetServicePathForIpConf(), node), &ed, time.Now().Unix())
	if err != nil {
		panic(err)
	}
	go sr.ListenLeaseRespChan()

	go func() {
		time.Sleep(5 * time.Second)
		sr.Close()
	}()

	// fmt.Println("已经下线")
}
