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
				setTcpConfig(conn) // KeepAlive
				if err != nil {
					if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
						fmt.Errorf("accept temp err: %v", nerr)
						continue
					}
					fmt.Errorf("accept err: %v", e)
				}
				// 新建一个链接，并派发conn.id
				c := NewConnection(conn)
				// 通过channel的方式
				e.addTask(c)
			}
		}()
	}
}

func (e *ePool) addTask(c *connection) {
	e.eChan <- c
}

func (e *ePool) startEPool() {
	// epoller的数量
	for i := 0; i < e.eSize; i++ {
		go e.startEProc()
	}
}

// 轮训器池 处理器
func (e *ePool) startEProc() {
	// 创建一个新epoller
	eper, err := newEpoller()
	if err != nil {
		panic(err)
	}
	// addTask事件，当前epoller中添加conn
	go func() {
		for {
			select {
			case <-e.done:
				return
			case conn := <-e.eChan:
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
	// 当前epoller等待200ms的conn，然后一起处理，最多100个（epoll_wait_queue_size）
	for {
		select {
		case <-e.done:
			return
		default:
			// 200ms一次轮训避免忙轮训
			connections, err := eper.wait(200)
			// 非空，且不是系统中断，则错误
			if err != nil && err != syscall.EINTR {
				fmt.Printf("failed to epoll wait %v\n", err)
				continue
			}
			for _, conn := range connections {
				if conn == nil {
					break
				}
				// runProc函数
				e.f(conn, eper)
			}
		}
	}
}

// 对象，轮训器
type epoller struct {
	fd            int      // epoll实例的文件描述符
	fdToConnTable sync.Map // 文件描述符映射到链接
}

// 创建epoller实例
func newEpoller() (*epoller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoller{fd: fd}, nil
}

func (e *epoller) add(conn *connection) error {
	addTcpNum()
	fmt.Printf("tcpNum:%d\n", getTcpNum())

	fd := conn.fd
	// 将链接的文件描述符fd添加到实例epoller中，并制定事件“unix.EPOLLIN”，“unix.EPOLLHUP”
	// unix.EPOLLIN：表示文件描述符上有数据可读
	// unix.EPOLLHUP：表示挂起（hang up）事件
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	// 存储epoller：fd -> conn
	e.fdToConnTable.Store(conn.fd, conn)
	// 大表 epool：conn.id -> conn
	ep.tables.Store(conn.id, conn)
	// 反向绑定
	conn.BindEpoller(e)
	return nil
}

func (e *epoller) remove(conn *connection) error {
	subTcpNum()
	fmt.Printf("tcpNum:%d\n", getTcpNum())

	fd := conn.fd
	// epoller实例中移除conn（通过conn.fd）
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	// 从两张表中删除链接信息
	ep.tables.Delete(conn.id)
	e.fdToConnTable.Delete(conn.fd)
	return nil
}

// 等待epoll实例上的事件，持续指定的时间，防止忙轮训
func (e *epoller) wait(msec int) ([]*connection, error) {
	events := make([]unix.EpollEvent, config.GetGatewayEpollWaitQueueSize())
	n, err := unix.EpollWait(e.fd, events, msec)
	if err != nil {
		return nil, err
	}
	var connections []*connection
	for i := 0; i < n; i++ {
		if conn, ok := e.fdToConnTable.Load(int(events[i].Fd)); ok {
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
