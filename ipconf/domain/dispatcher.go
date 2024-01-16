package domain

import (
	"sort"
	"sync"

	"github.com/kuan525/tiger/ipconf/source"
)

type Dispatcher struct {
	// 通过source中的event的channel修改，后续从这个里面取数据
	candidateTable map[string]*Endport
	sync.RWMutex
}

var dp *Dispatcher

func Init() {
	dp = &Dispatcher{}
	dp.candidateTable = make(map[string]*Endport)
	go func() {
		// 这里获取的是source中存储event的channel，在调度层处理
		for event := range source.EventChan() {
			switch event.Type {
			case source.AddNodeEvent:
				dp.addNode(event)
			case source.DelNodeEvent:
				dp.delNode(event)
			}
		}
	}()
}

func Dispatch(ctx *IpConfContext) []*Endport {
	// step1 获取候选endport
	eds := dp.getCandidateEndport(ctx)
	// step2 逐一计算得分
	for _, ed := range eds {
		ed.CalculateScore(ctx)
	}
	// step3 全局排序，动静结合的排序策略
	sort.Slice(eds, func(i, j int) bool {
		// 优先基于活跃分数进行排序
		if eds[i].ActiveSorce > eds[j].ActiveSorce {
			return true
		} else if eds[i].ActiveSorce < eds[j].ActiveSorce {
			return false
		}
		// 如果活跃分相同，则使用静态分排序
		return eds[i].StaticSorce > eds[j].StaticSorce
	})
	return eds
}

func (dp *Dispatcher) getCandidateEndport(ctx *IpConfContext) []*Endport {
	dp.RLock()
	defer dp.RUnlock()
	// 这里先将map中的所有都拷贝出来，后续再操作，这里网关机器数量才是瓶颈
	candidateList := make([]*Endport, 0, len(dp.candidateTable))
	for _, ed := range dp.candidateTable {
		candidateList = append(candidateList, ed)
	}
	return candidateList
}

func (dp *Dispatcher) delNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	delete(dp.candidateTable, event.Key())
}

func (dp *Dispatcher) addNode(event *source.Event) {
	dp.Lock()
	defer dp.Unlock()
	var (
		ed *Endport
		ok bool
	)
	if ed, ok = dp.candidateTable[event.Key()]; !ok { // 不存在
		// soucre:event类型变为endport类型
		ed = NewEndport(event.IP, event.Port)
		dp.candidateTable[event.Key()] = ed
	}

	ed.UpdateStat(&Stat{
		ConnectNum:   event.ConnectNum,
		MessageBytes: event.MessageBytes,
	})
}
