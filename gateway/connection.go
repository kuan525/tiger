package gateway

import (
	"errors"
	"net"
	"sync"
	"time"
)

var node *ConnIDGenerater

const (
	version      = uint64(0) // 版本控制
	sequenceBits = uint64(16)
	maxSequence  = int64(-1) ^ (int64(-1) << sequenceBits)
	timeLeft     = uint8(16) // 时间戳向左偏移量
	versionLeft  = uint8(63) // 左移到最高位
	// 2020-05-20 08:00:00 +0800 CST
	twepoch = int64(1589923200000) // 常量时间戳（毫秒）
)

// ConnId生成器
type ConnIDGenerater struct {
	mu        sync.Mutex
	LastStamp int64 // 记录上一次ID的时间戳
	Sequence  int64 // 当前毫秒已经生成的ID序列号（从0开始累加）1毫秒内最多生成2^16个ID
}

type connection struct {
	id   uint64 // 进程级别的生命周期
	fd   int
	e    *epoller
	conn *net.TCPConn
}

func init() {
	node = &ConnIDGenerater{}
}

func (c *ConnIDGenerater) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func NewConnection(conn *net.TCPConn) *connection {
	var id uint64
	var err error
	if id, err = node.NextID(); err != nil {
		panic(err) // 在线服务需要解决这个问题，报错而不能panic
	}
	return &connection{
		id:   id,
		fd:   socketFD(conn),
		conn: conn,
	}
}

func (c *connection) Close() {
	ep.tables.Delete(c.id)
	if c.e != nil {
		c.e.fdToConnTable.Delete(c.fd)
	}
	err := c.conn.Close()
	panic(err)
}

func (c *connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *connection) BindEpoller(e *epoller) {
	c.e = e
}

// 这里的锁会自旋，不会多么影响性能，主要是临界区小
func (w *ConnIDGenerater) NextID() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.nextID()
}

func (w *ConnIDGenerater) nextID() (uint64, error) {
	timeStamp := w.getMilliSeconds()
	if timeStamp < w.LastStamp {
		return 0, errors.New("time is moving backwards, waiting untill")
	}
	if w.LastStamp == timeStamp {
		w.Sequence = w.Sequence + 1
		if w.Sequence == maxSequence { // 如果这里发生溢出，就等到下一个毫秒再分配，这样就一定不会出现重复
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
			w.Sequence = 0
		}
	} else {
		w.Sequence = 0
	}
	w.LastStamp = timeStamp
	// 减法可以压缩一下时间戳
	id := ((timeStamp - twepoch) << timeLeft) | w.Sequence
	connID := uint64(id) | (version << versionLeft)
	return connID, nil
}
