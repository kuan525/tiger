package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/kuan525/tiger/common/tgrpc/discov"

	"github.com/stretchr/testify/assert"
)

func TestNewETCDRegister(t *testing.T) {
	_, err := NewETCDRegister()
	assert.Nil(t, err)
}

func TGestRegister_Register(t *testing.T) {
	register, _ := NewETCDRegister()
	service := &discov.Service{
		Name: "test",
		Endpoints: []*discov.Endpoint{
			{
				ServiceName: "test",
				IP:          "127.0.0.1",
				Port:        9557,
				Weight:      100,
				Enable:      true,
			},
		},
	}
	register.Register(context.TODO(), service)
	time.Sleep(2 * time.Second)
	registerService := register.GetService(context.TODO(), "test")
	assert.Equal(t, *service.Endpoints[0], *registerService.Endpoints[0])
}
