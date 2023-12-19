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

// 新建服务发现
// source的主要逻辑，这里是一个协程处理，是将etcd中的变更传入到eventChan中，
// 等待dispatcher通过eventChan去操作dis到map，也就是对应修改/删除操作。
// 具体的：这里不需要接口操作，所有的数据通过etcd watch发现，得到的数据是etcd的k-v结构
// 利用set和del去操作这个k-v，一个k-v等于是一个gateway，操作的具体行为是插入eventChan，source只需要做好这个就好了。
func DataHandler(ctx *context.Context) {
	dis := discovery.NewServiceDiscovery(ctx)
	defer dis.Close()

	// 修改/删除函数 - 将这两个函数传入服务发现的逻辑中，具体点：将在etch中扫描到的k-v通过下面两个函数传入到eventChan中
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
