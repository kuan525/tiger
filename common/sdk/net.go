package sdk

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/kuan525/tiger/common/tcp"
)

type connect struct {
	sendChan, recvChan chan *Message
	conn               *net.TCPConn
}

func newConnet(ip net.IP, port int) *connect {
	clientConn := &connect{
		sendChan: make(chan *Message),
		recvChan: make(chan *Message),
	}
	addr := &net.TCPAddr{IP: ip, Port: port}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Printf("DialTCP.err=%+v", err)
		return nil
	}
	clientConn.conn = conn
	go func() {
		for {
			data, err := tcp.ReadData(conn)
			if err != nil {
				fmt.Printf("ReadData err:%+v", err)
			}
			msg := &Message{}
			err = json.Unmarshal(data, msg)
			if err != nil {
				panic(err)
			}
			clientConn.recvChan <- msg
		}
	}()
	return clientConn
}

func (c *connect) send(data *Message) {
	// 直接发送给接收方
	bytes, _ := json.Marshal(data)
	datapkg := tcp.DataPkg{
		Data: bytes,
		Len:  uint32(len(bytes)),
	}
	val := datapkg.Marshal()
	c.conn.Write(val)

	// c.recvChan <- data
}

func (c *connect) recv() <-chan *Message {
	return c.recvChan
}

func (c *connect) close() {

}
