package server

import (
	"context"

	ttrace "github.com/kuan525/tiger/common/tgrpc/trace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TraceUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md := metadata.MD{}
		header, ok := metadata.FromIncomingContext(ctx)
		if ok {
			md = header.Copy()
		}

		spanCtx := ttrace.Extract(ctx, otel.GetTextMapPropagator(), &md)
		tr := otel.Tracer(ttrace.TraceName)
		name, attrs := ttrace.BuildSpan(info.FullMethod, ttrace.PeerFromCtx(ctx))

		ctx, span := tr.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), name, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attrs...))
		defer span.End()

		resp, err = handler(ctx, req)
		if err != nil {
			s, ok := status.FromError(err)
			if ok {
				span.SetStatus(codes.Error, s.Message())
				span.SetAttributes(ttrace.StatusCodeAttr(s.Code()))
			} else {
				span.SetStatus(codes.Error, s.Message())
			}
			return nil, nil
		}
		return resp, nil
	}
}
