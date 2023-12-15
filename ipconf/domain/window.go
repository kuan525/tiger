package domain

type stateWindow struct {
	stateQueue []*Stat
	statChan   chan *Stat
	sumStat    *Stat
	idx        int64
}

const windowSize = 5

func newStateWindow() *stateWindow {
	return &stateWindow{
		stateQueue: make([]*Stat, windowSize),
		statChan:   make(chan *Stat),
		sumStat:    &Stat{},
	}
}

func (sw *stateWindow) getStat() *Stat {
	res := sw.sumStat.Clone()
	res.Avg(windowSize)
	return res
}

func (sw *stateWindow) appendStat(s *Stat) {
	sw.sumStat.Sub(sw.stateQueue[sw.idx%windowSize])
	sw.stateQueue[sw.idx%windowSize] = s
	sw.sumStat.Add(s)
	sw.idx++
}
