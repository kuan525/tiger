package domain

import (
	"sync/atomic"
	"unsafe"
)

type Endport struct {
	IP          string       `json:"ip"`
	Port        string       `json:"port"`
	ActiveSorce float64      `json:"-"`
	StaticSorce float64      `json:"-"`
	Stats       *Stat        `json:"-"`
	window      *stateWindow `json:"-"`
}

func NewEndport(ip, port string) *Endport {
	ed := &Endport{
		IP:   ip,
		Port: port,
	}
	ed.window = newStateWindow()
	ed.Stats = ed.window.getStat()
	go func() {
		// 内存泄漏：机器下线之后，该协程未停止
		for stat := range ed.window.statChan {
			ed.window.appendStat(stat)     // 更新window中
			newStat := ed.window.getStat() // 通过window中更新stat
			// 新状态替换旧状态
			// 旧协程在处理stats，同时通知ed下线，map的delete删除，马上ed上线，
			// 这时候会出现stats处理冲突
			atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(ed.Stats)), unsafe.Pointer(newStat))
		}
	}()
	return ed
}

func (ed *Endport) UpdateStat(s *Stat) {
	ed.window.statChan <- s
}

func (ed *Endport) CalculateScore(ctx *IpConfContext) {
	// 如果stats字端是空的，则直接使用上一次计算的结果，此次不更新
	// 正常情况下stats非空，每一次更新状态的时候都会更新
	if ed.Stats != nil {
		ed.ActiveSorce = ed.Stats.CalculateActiveSorce() // GB
		ed.StaticSorce = ed.Stats.CalculateStaticSorce() // 个
	}
}
