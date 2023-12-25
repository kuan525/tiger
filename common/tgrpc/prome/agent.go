package prome

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var once sync.Once

// 开启prometheus
func StartAgent(host string, port int) {
	go func() {
		once.Do(func() {
			http.Handle("/", promhttp.Handler())
			addr := fmt.Sprintf("%s:%d", host, port)
			logger.Infof("Starting prometheus agent at %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Error(err)
			}
		})
	}()
}
