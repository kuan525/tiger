package server

import (
	"context"
	"time"

	"github.com/kuan525/tiger/common/tgrpc/prome"
	"github.com/kuan525/tiger/common/tgrpc/util"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const nameSpace = "tgrpc_server"

var (
	serverHandleCounter = prome.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: nameSpace,
			Subsystem: "req",
			Name:      "client_handle_total",
		},
		[]string{"method", "server", "code", "ip"},
	)

	serverHandleHistogram = prome.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: nameSpace,
			Subsystem: "req",
			Name:      "client_handle_seconds",
		},
		[]string{"method", "server", "ip"},
	)
)

// 记录请求开始的时间，调用原始的处理函数，然后记录请求处理的时间和错误代码
func MetricUnaryServerInterceptor(serverName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		beg := time.Now()
		resp, err = handler(ctx, req)

		code := status.Code(err)
		serverHandleCounter.WithLabelValues(info.FullMethod, serverName, code.String(), util.ExternaIP()).Inc()
		serverHandleHistogram.WithLabelValues(info.FullMethod, serverName, util.ExternaIP()).Observe(time.Since(beg).Seconds())
		return
	}
}
