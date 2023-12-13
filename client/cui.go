package client

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gookit/color"
	"github.com/kuan525/tiger/client/sdk"
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
	fmt.Fprintf(v, "%v:%v%v\n", color.FgGreen.Text(self.Name), self.Sep,
		color.FgYellow.Text(self.Msg))
	return nil
}

func viewPrint(g *gocui.Gui, name, msg string, newline bool) {

}
func doRecv(g *gocui.Gui) {

}
func quit(g *gocui.Gui, v *gocui.View) error {

	return nil
}

func doSay(g *gocui.Gui, cv *gocui.View) {

}

func viewUpdate(g *gocui.Gui, cv *gocui.View) error {

}

func vierUpScroll(g *gocui.Gui, cv *gocui.View) error {

}

func viewDownScroll(g *gocui.Gui, cv *gocui.View) error {

}

func viewOutput(g *gocui.Gui, x0, y0, x1, y1 int) error {

}

func viewInput(g *gocui.Gui, x0, y0, x1, y1 int) error {

}

func viewHead(g *gocui.Gui, x0, y0, x1, y1 int) error {

}

func layout(g *gocui.Gui) error {

}

func pasteUP(g *gocui.Gui, cv *gocui.View) error {

}

func pasteDown(g *gocui.Gui, cv *gocui.View) error {

}

func RunMain() {

}
