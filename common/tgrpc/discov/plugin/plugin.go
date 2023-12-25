package plugin

import (
	"errors"
	"fmt"

	"github.com/kuan525/tiger/common/tgrpc/config"
	"github.com/kuan525/tiger/common/tgrpc/discov"
	"github.com/kuan525/tiger/common/tgrpc/discov/etcd"
)

func GetDiscovInstance() (discov.Discovery, error) {
	name := config.GetDiscovName()
	switch name {
	case "etcd":
		return etcd.NewETCDRegister(etcd.WithEndpoints(config.GetDiscovEndpoints()))
	}
	return nil, errors.New(fmt.Sprintf("not exist plugin:%s", name))
}
