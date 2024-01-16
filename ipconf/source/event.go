package source

import (
	"fmt"

	"github.com/kuan525/tiger/common/discovery"
)

type EventType string

type Event struct {
	Type         EventType
	IP           string
	Port         string
	ConnectNum   float64
	MessageBytes float64
}

const (
	AddNodeEvent EventType = "addNode"
	DelNodeEvent EventType = "delNode"
)

var eventChan chan *Event

func EventChan() <-chan *Event {
	return eventChan
}

func (nd *Event) Key() string {
	return fmt.Sprintf("%s:%s", nd.IP, nd.Port)
}

// 一个event，对应一个endpoint的操作，增加/删除/更新
func NewEvent(ed *discovery.EndpointInfo) *Event {
	if ed == nil || ed.MetaData == nil {
		return nil
	}
	var connNum, msgBytes float64
	if data, ok := ed.MetaData["connect_num"]; ok {
		connNum = data.(float64) // 如果出错，此处应该panic暴露错误
	}
	if data, ok := ed.MetaData["message_bytes"]; ok {
		msgBytes = data.(float64) // 如果出错，此处应该panic暴露错误
	}
	return &Event{
		// Type:         AddNodeEvent,
		IP:           ed.IP,
		Port:         ed.Port,
		ConnectNum:   connNum,
		MessageBytes: msgBytes,
	}
}
