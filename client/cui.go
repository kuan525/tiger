package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/gookit/color"
	"github.com/kuan525/tiger/common/config"
	"github.com/kuan525/tiger/common/sdk"
	"github.com/rocket049/gocui"
)

var (
	buf     string
	chat    *sdk.Chat
	step    int
	verbose bool
	pos     int
)

type VOT struct {
	Name string
	Msg  string
	Sep  string
}

func init() {
	// r = rand.New(rand.NewSource(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())
	verbose = true
	step = 1
}

func setHeadText(g *gocui.Gui, msg string) {
	v, err := g.View("head")
	if err == nil {
		v.Clear()
		fmt.Fprint(v, color.FgGreen.Text(msg))
	}
}

func (self VOT) Show(g *gocui.Gui) error {
	v, err := g.View("out")
	if err != nil {
		// log.Println("No output view")
		return nil
	}
	// 类似写入缓冲区，等监听函数执行完再刷新
	fmt.Fprintf(v, "%v:%v%v\n", color.FgGreen.Text(self.Name), self.Sep, color.FgYellow.Text(self.Msg))
	return nil
}

func viewPrint(g *gocui.Gui, name, msg string, newline bool) {
	var out VOT
	out.Name, out.Msg = name, msg
	if newline {
		out.Sep = "\n"
	} else {
		out.Sep = " "
	}
	// out.Show(g) // 仅仅为了展示，保序
	g.Update(out.Show) // 使用channel安全的更新，但是无法保证顺序
}

func doRecv(g *gocui.Gui) {
	recvChannel := chat.Recv()
	for msg := range recvChannel {
		if msg != nil {
			switch msg.Type {
			case sdk.MsgTypeText:
				viewPrint(g, msg.Name, msg.Content, false)
			case sdk.MsgTypeAck:
				// TODO 默认不处理
			}
		}
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	chat.Close()
	ov, _ := g.View("out")
	buf = ov.Buffer()
	g.Close()
	return gocui.ErrQuit
}

func doSay(g *gocui.Gui, cv *gocui.View) {
	v, err := g.View("out")
	if cv != nil && err == nil {
		p := cv.ReadEditor()
		if p != nil {
			msg := &sdk.Message{
				Type:       sdk.MsgTypeText,
				Name:       "KuanBot",
				FormUserID: "123213",
				ToUserID:   "222222",
				Content:    string(p),
			}
			idKey := fmt.Sprintf("%d", chat.GetCurClientID())
			viewPrint(g, "me:"+idKey, msg.Content, false)
			chat.Send(msg)
		}
		v.Autoscroll = true
	}
}

// 整个完成之后才会渲染到界面
func viewUpdate(g *gocui.Gui, cv *gocui.View) error {
	doSay(g, cv)
	l := len(cv.Buffer())
	cv.MoveCursor(0-l, 0, true)
	cv.Clear()
	return nil
}

func viewUpScroll(g *gocui.Gui, cv *gocui.View) error {
	v, err := g.View("out")
	v.Autoscroll = false
	ox, oy := v.Origin()
	if err == nil {
		v.SetOrigin(ox, oy-1)
	}
	return nil
}

func viewDownScroll(g *gocui.Gui, cv *gocui.View) error {
	v, err := g.View("out")
	_, y := v.Size()
	ox, oy := v.Origin()
	lnum := len(v.BufferLines())
	if err == nil {
		if oy > lnum-y-1 {
			v.Autoscroll = true
		} else {
			v.SetOrigin(ox, oy+1)
		}
	}
	return nil
}

func viewOutput(g *gocui.Gui, x0, y0, x1, y1 int) error {
	v, err := g.SetView("out", x0, y0, x1, y1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Overwrite = false
		v.Autoscroll = true
		v.SelBgColor = gocui.ColorRed
		v.Title = "Messages"
	}
	return nil
}

func viewInput(g *gocui.Gui, x0, y0, x1, y1 int) error {
	if v, err := g.SetView("main", x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		//当 err == gocui.ErrUnknownView 时运行
		v.Editable = true
		v.Wrap = true
		v.Overwrite = false
		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	return nil
}

func viewHead(g *gocui.Gui, x0, y0, x1, y1 int) error {
	if v, err := g.SetView("head", x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = false
		v.Overwrite = true
		msg := "tiger: IM系统聊天对话框 【由于gocui使用channel异步处理任务，所以不保序～】"
		setHeadText(g, msg)
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if err := viewHead(g, 1, 1, maxX-1, 3); err != nil {
		return err
	}
	if err := viewOutput(g, 1, 4, maxX-1, maxY-4); err != nil {
		return err
	}
	if err := viewInput(g, 1, maxY-3, maxX-1, maxY-1); err != nil {
		return err
	}
	return nil
}

func pasteUP(g *gocui.Gui, cv *gocui.View) error {
	v, err := g.View("out")
	if err != nil {
		fmt.Fprintf(cv, "error:%s", err)
		return nil
	}
	bls := v.BufferLines()
	lnum := len(bls)
	if pos < lnum-1 {
		pos++
	}
	cv.Clear()
	fmt.Fprintf(cv, "%s", bls[lnum-pos-1])
	return nil
}

func pasteDown(g *gocui.Gui, cv *gocui.View) error {
	v, err := g.View("out")
	if err != nil {
		fmt.Fprintf(cv, "error:%s", err)
		return nil
	}
	if pos > 0 {
		pos--
	}
	bls := v.BufferLines()
	lnum := len(bls)
	cv.Clear()
	fmt.Fprintf(cv, "%s", bls[lnum-pos-1])
	return nil
}

func RunMain(path string) {
	config.Init(path)
	// step1 创建chat的核心对象
	chat = sdk.NewChat(net.ParseIP("0.0.0.0"), config.GetGatewayTCPServerPort(), "kuan", "12312321", "2131")
	// step2 创建GUI图层对象并进行参与与回调函数的配置
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	g.Cursor = true
	g.Mouse = false
	g.ASCII = false

	// 设置编排函数
	g.SetManagerFunc(layout)

	// 注册回调事件
	if err := g.SetKeybinding("main", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone, viewUpdate); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyPgup, gocui.ModNone, viewUpScroll); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyPgdn, gocui.ModNone, viewDownScroll); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, pasteDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, pasteUP); err != nil {
		log.Panicln(err)
	}

	go func() {
		time.Sleep(10 * time.Second)
		// 模拟一次断线
		chat.ReConn()
	}()

	// 启动消费函数
	go doRecv(g)
	if err := g.MainLoop(); err != nil {
		log.Println(err)
	}
	ioutil.WriteFile("chat.log", []byte(buf), 0644)
}
