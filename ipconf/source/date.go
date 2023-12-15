package source

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/discovery"
)

func Init() {
	eventChan = make(chan *Event)
	ctx := context.Background()

	go DataHandler(&ctx)

	if config.IsDebug() {
		ctx := context.Background()
		testServiceRegister(&ctx, "7896", "node1")
		testServiceRegister(&ctx, "7897", "node2")
		testServiceRegister(&ctx, "7898", "node3")
	}
}

func DataHandler(ctx *context.Context) {
	dis := discovery.NewServiceDiscovery(ctx)
	defer dis.Close()

	setFunc := func(key, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); ed != nil {
				event.Type = AddNodeEvent
				eventChan <- event
			}
		} else {
			logger.CtxErrorf(*ctx, "DataHandler.setFunc.err :%s", err.Error())
		}
	}
	delFunc := func(key, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); ed != nil {
				event.Type = DelNodeEvent
				eventChan <- event
			}
		} else {
			logger.CtxErrorf(*ctx, "dataHandler,delFunc,err :%s", err.Error())
		}
	}
	err := dis.WatchService(config.GetServicePathForIpConf(), setFunc, delFunc)
	if err != nil {
		panic(err)
	}
}
