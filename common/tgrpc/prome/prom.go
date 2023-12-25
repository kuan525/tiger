package prome

import "github.com/prometheus/client_golang/prometheus"

// 计数器向量
func NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	counterVec := prometheus.NewCounterVec(opts, labelNames)
	prometheus.MustRegister(counterVec)
	return counterVec
}

// 直方图向量
func NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	histogramVec := prometheus.NewHistogramVec(opts, labelNames)
	prometheus.MustRegister(histogramVec)
	return histogramVec
}
