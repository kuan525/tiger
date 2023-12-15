package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/kuan525/tiger/common/config"
)

func TestServiceDiscovery(t *testing.T) {
	config.Init("../../tiger.yaml")
	ctx := context.Background()
	ser := NewServiceDiscovery(&ctx)
	defer ser.Close()

	go ser.WatchService("/web/", func(key, value string) {}, func(key, value string) {})
	go ser.WatchService("/gRPC/", func(key, value string) {}, func(key, value string) {})

	time.Sleep(time.Second * 30)
}
