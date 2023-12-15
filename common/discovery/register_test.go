package discovery

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/kuan525/tiger/common/config"
)

func TestServiceRegister(t *testing.T) {
	config.Init("../../tiger.yaml")
	ctx := context.Background()
	ser, err := NewServiceRegister(&ctx, "/web/node1", &EndpointInfo{
		IP:   "127.0.0.1",
		Port: "9999",
	}, 5)
	if err != nil {
		log.Fatalln(err)
	}
	defer ser.Close()

	// 监听续租相应chan
	go ser.ListenLeaseRespChan()

	time.Sleep(time.Second * 30)
}
