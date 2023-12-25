package trace

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

const (
	TraceName = "tiger-trace"
)

type metadataSupplier struct {
	metadata *metadata.MD
}

func (s *metadataSupplier) Get(key string) string {
	values := s.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (s *metadataSupplier) Set(key, value string) {
	s.metadata.Set(key, value)
}

func (s *metadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.metadata))
	for key := range *s.metadata {
		out = append(out, key)
	}
	return out
}

func Inject(ctx context.Context, p propagation.TextMapPropagator, m *metadata.MD) {
	p.Inject(ctx, &metadataSupplier{
		metadata: m,
	})
}

func Extract(ctx context.Context, p propagation.TextMapPropagator, m *metadata.MD) sdktrace.SpanContext {
	ctx = p.Extract(ctx, &metadataSupplier{
		metadata: m,
	})
	return sdktrace.SpanContextFromContext(ctx)
}
