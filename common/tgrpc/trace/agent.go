package trace

import (
	"context"
	"sync"

	"github.com/kuan525/tiger/common/tgrpc/config"

	"github.com/bytedance/gopkg/util/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var (
	tp   *tracesdk.TracerProvider
	once sync.Once
)

// 开启trace collector
func StartAgent() {
	once.Do(func() {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.GetTraceCollectionUrl())))
		if err != nil {
			logger.Errorf("trace start agent err:%s", err.Error())
			return
		}

		tp = tracesdk.NewTracerProvider(
			tracesdk.WithSampler(tracesdk.TraceIDRatioBased(config.GetTraceSampler())),
			tracesdk.WithBatcher(exp),
			tracesdk.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.GetTraceServiceName()),
			)),
		)

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{}))
	})
}

// 关闭trace collector，在服务停止前调用stopAgent，不然可能造成trace数据的丢失
func StopAgent() {
	_ = tp.Shutdown(context.TODO())
}
