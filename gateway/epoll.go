package gateway

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/kuan525/tiger/common/config"
	"golang.org/x/sys/unix"
)

// 全局对象
var (
	ep     *ePool // epoll池
	tcpNum int32  // 当前服务器允许接入的最大tcp连接数
)

type ePool struct {
	eChan  chan *connection
	tables sync.Map
	eSize  int
	done   chan struct{}

	ln *net.TCPListener
	f  func(c *connection, ep *epoller)
}

func initEpoll(ln *net.TCPListener, f func(c *connection, ep *epoller)) {
	setLimit()
	ep = newEpool(ln, f)
	ep.createAcceptProcess()
	ep.startEPool()
}

// 设置go 进程打开文件数的限制
func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	log.Printf("set cur limit: %d", rLimit.Cur)
}

func newEpool(ln *net.TCPListener, cb func(c *connection, ep *epoller)) *ePool {
	return &ePool{
		eChan:  make(chan *connection, config.GetGatewayEpollerChanNum()),
		done:   make(chan struct{}),
		eSize:  config.GetGatewayEpollerNum(),
		tables: sync.Map{},
		ln:     ln,
		f:      cb,
	}
}

// 创建一个专门处理accept事件的协程，与当前cpu的核数对应，能发挥最大功效
func (e *ePool) createAcceptProcess() {
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				conn, err := e.ln.AcceptTCP()
				// 限流熔断
				if !checkTcp() {
					_ = conn.Close()
					continue
				}
				setTcpConfig(conn)
				if err != nil {
					if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
						fmt.Errorf("accept temp err: %v", nerr)
						continue
					}
					fmt.Errorf("accept err: %v", e)
				}
				c := connection{
					conn: conn,
					fd:   socketFD(conn),
				}
				e.addTask(&c)
			}
		}()
	}
}

func (e *ePool) addTask(c *connection) {
	e.eChan <- c
}

func (e *ePool) startEPool() {
	for i := 0; i < e.eSize; i++ {
		go e.startEProc()
	}
}

// 轮训器池 处理器
func (e *ePool) startEProc() {
	eper, err := newEpoller()
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case <-e.done:
				return
			case conn := <-e.eChan:
				addTcpNum()
				fmt.Printf("tcpNum:%d\n", getTcpNum())
				if err := eper.add(conn); err != nil {
					fmt.Printf("failed to add connection %v\n", err)
					conn.Close() //登陆未成功直接关闭连接
					continue
				}
				fmt.Printf("EpollerPoll new connection[%v] tcpSize:%d\n", conn.RemoteAddr(), getTcpNum())
			}
		}
	}()
	// 轮训器在这里轮训等待，当有wait发生时则调用回调函数去处理
	for {
		select {
		case <-e.done:
			return
		default:
			connections, err := eper.wait(200) // 200ms一次轮训避免忙轮训

			if err != nil && err != syscall.EINTR {
				fmt.Printf("failed to epoll wait %v\n", err)
				continue
			}
			for _, conn := range connections {
				if conn == nil {
					break
				}
				e.f(conn, eper)
			}
		}
	}
}

// 对象，轮训器
type epoller struct {
	fd int
}

func newEpoller() (*epoller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoller{fd: fd}, nil
}

// TODO: 默认水平出发模式，可采用非阻塞FD，优化边沿出发模式
func (e *epoller) add(conn *connection) error {
	fd := conn.fd
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	ep.tables.Store(fd, conn)
	return nil
}

func (e *epoller) remove(conn *connection) error {
	subTcpNum()
	fd := conn.fd
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	ep.tables.Delete(fd)
	return nil
}

func (e *epoller) wait(msec int) ([]*connection, error) {
	events := make([]unix.EpollEvent, config.GetGatewayEpollWaitQueueSize())
	n, err := unix.EpollWait(e.fd, events, msec)
	if err != nil {
		return nil, err
	}
	var connections []*connection
	for i := 0; i < n; i++ {
		if conn, ok := ep.tables.Load(int(events[i].Fd)); ok {
			connections = append(connections, conn.(*connection))
		}
	}
	return connections, nil
}

func socketFD(conn *net.TCPConn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(*conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func addTcpNum() {
	atomic.AddInt32(&tcpNum, 1)
}
func getTcpNum() int32 {
	return atomic.LoadInt32(&tcpNum)
}
func subTcpNum() {
	atomic.AddInt32(&tcpNum, -1)
}
func checkTcp() bool {
	num := getTcpNum()
	maxTcpNum := config.GetGatewayMaxTcpNum()
	return num <= maxTcpNum
}
func setTcpConfig(c *net.TCPConn) {
	_ = c.SetKeepAlive(true)
}
