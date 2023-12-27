package logger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kuan525/tiger/common/config"
	ttrace "github.com/kuan525/tiger/common/tgrpc/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestLogger(t *testing.T) {
	config.Init("../../tiger.yaml")
	NewLogger(WithLogDir("../../logs"))
	InfoCtx(context.Background(), "info test")
	DebugCtx(context.Background(), "debug test")
	WarnCtx(context.Background(), "warn test")
	ErrorCtx(context.Background(), "error test")
	time.Sleep(1 * time.Second)
}

func TestTraceLog(t *testing.T) {
	config.Init("../../tiger.yaml")
	NewLogger(WithLogDir("../../logs"))
	ttrace.StartAgent()
	defer ttrace.StopAgent()

	tr := otel.GetTracerProvider().Tracer(ttrace.TraceName)
	ctx, span := tr.Start(context.Background(), "logger-trace", trace.WithAttributes(), trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	InfoCtx(ctx, "Info text")
	ErrorCtx(ctx, "Error text")
}

func TestGetTraceID(t *testing.T) {
	fmt.Println("GetTraceID:", GetTraceID(context.Background()))
}
