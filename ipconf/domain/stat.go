package domain

import "math"

// 数值代表的是，此endpoint对应的机器其自身剩余的资源指标
type Stat struct {
	MessageBytes float64 // 活跃分 gateway 每秒钟收发字节数的 剩余值
	ConnectNum   float64 // 静态分 gateway 总体持有的长连接数量的 剩余值
}

func (s *Stat) CalculateActiveSorce() float64 {
	return getGB(s.MessageBytes)
}

func (s *Stat) Avg(num float64) {
	s.ConnectNum /= num
	s.MessageBytes /= num
}

func (s *Stat) Clone() *Stat {
	return &Stat{
		MessageBytes: s.MessageBytes,
		ConnectNum:   s.ConnectNum,
	}
}

func (s *Stat) Add(st *Stat) {
	if st == nil {
		return
	}
	s.ConnectNum += st.ConnectNum
	s.MessageBytes += st.MessageBytes
}

func (s *Stat) Sub(st *Stat) {
	if st == nil {
		return
	}
	s.ConnectNum -= st.ConnectNum
	s.MessageBytes -= st.MessageBytes
}

func getGB(m float64) float64 {
	return decimal(m / (1 << 30))
}

func decimal(value float64) float64 {
	return math.Trunc(value*1e2+0.5) * 1e-2
}

func (s *Stat) CalculateStaticSorce() float64 {
	return s.ConnectNum
}
