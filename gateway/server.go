package gateway

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/tcp"
)

// 启动网关服务 [配置文件路径]
func RunMain(path string) {
	config.Init(path)
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: config.GetGatewayTCPServerPort()})
	if err != nil {
		log.Fatalf("StartTCPEPollServer err:%s", err.Error())
		panic(err)
	}
	initWorkPool()
	initEpoll(ln, runProc)
	fmt.Println("-------------im gateway stated-------------")
	// 主线程不退出
	select {}
}

func runProc(c *connection, eper *epoller) {
	// step1: 读取一个完整的包
	dataBuf, err := tcp.ReadData(c.conn)
	if err != nil {
		// 如果读取conn时发现连接关闭，则直接关闭端口连接
		if errors.Is(err, io.EOF) {
			eper.remove(c)
		}
		return
	}
	err = wPool.Submit(func() {
		// step2:交给state server rpc处理
		bytes := tcp.DataPkg{
			Len:  uint32(len(dataBuf)),
			Data: dataBuf,
		}
		tcp.SendData(c.conn, bytes.Marshal())
	})
	if err != nil {
		fmt.Errorf("runProc:err:%+v\n", err.Error())
	}
}
