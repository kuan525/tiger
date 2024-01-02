package sdk

import (
	"net"
	"time"

	"github.com/kuan525/tiger/common/idl/message"
	"github.com/kuan525/tiger/common/tcp"
	"google.golang.org/protobuf/proto"
)

const (
	MsgTypeText      = "text"
	MsgTypeAck       = "ack"
	MsgTypeReConn    = "reConn"
	MsgTypeHeartbeat = "heartbeat"
	MsgLogin         = "loginMsg"
)

type Chat struct {
	Nick      string
	UserID    string
	SessionId string
	conn      *connect
	closeChan chan struct{}
}

type Message struct {
	Type       string
	Name       string
	FormUserID string
	ToUserID   string
	Content    string
	Session    string
}

func NewChat(ip net.IP, port int, nick, userID, sessionID string, connID uint64, isReConn bool) *Chat {
	chat := &Chat{
		Nick:      nick,
		UserID:    userID,
		SessionId: sessionID,
		conn:      newConnet(ip, port, connID),
		closeChan: make(chan struct{}),
	}
	go chat.loop()
	if isReConn {
		chat.reConn(connID)
	} else {
		chat.login()
	}
	go chat.heartbeat()
	return chat
}

func (chat *Chat) Send(msg *Message) {
	// chat.conn.send(msg)
	chat.conn.recvChan <- msg
}

func (chat *Chat) Recv() <-chan *Message {
	return chat.conn.recv()
}

func (chat *Chat) Close() {
	chat.conn.close()
	close(chat.closeChan)
	close(chat.conn.recvChan)
	close(chat.conn.sendChan)
}

func (chat *Chat) GetConnID() uint64 {
	return chat.conn.connID
}

func (chat *Chat) loop() {
	for {
		select {
		case <-chat.closeChan:
			return
		default:
			mc := &message.MsgCmd{}
			data, err := tcp.ReadData(chat.conn.conn)
			if err != nil {
				return
			}
			err = proto.Unmarshal(data, mc)
			if err != nil {
				panic(err)
			}
			var msg *Message
			switch message.CmdType_ACK {
			case message.CmdType_ACK:
				msg = handAckMsg(chat.conn, mc.Payload)
			}
			chat.conn.recvChan <- msg
		}
	}
}

func (chat *Chat) login() {
	loginMsg := message.LoginMsg{
		Head: &message.LoginMsgHead{DeviceID: 123},
	}
	payLoad, err := proto.Marshal(&loginMsg)
	if err != nil {
		panic(err)
	}
	chat.conn.send(message.CmdType_Login, payLoad)
}

func (chat *Chat) reConn(connID uint64) {
	reConn := message.ReConnMsg{
		Head: &message.ReConnMsgHead{ConnID: connID},
	}
	payLoad, err := proto.Marshal(&reConn)
	if err != nil {
		panic(err)
	}
	chat.conn.send(message.CmdType_ReConn, payLoad)
}

func (chat *Chat) heartbeat() {
	tc := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-chat.closeChan:
			return
		case <-tc.C:
			hearbeat := message.HeartbeatMsg{
				Head: &message.HeartbeatMsgHead{},
			}
			payLoad, err := proto.Marshal(&hearbeat)
			if err != nil {
				panic(err)
			}
			chat.conn.send(message.CmdType_Heartbeat, payLoad)
		}
	}
}
