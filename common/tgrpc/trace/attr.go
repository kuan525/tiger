package trace

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	gcodes "google.golang.org/grpc/codes"
)

const (
	// GRPCStatusCodeKey 是 gRPC 请求的数字状态代码的约定。
	GRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// RPCNameKey 是发送或接收的消息的名称。
	RPCNameKey = attribute.Key("name")
	// RPCMessageTypeKey 是发送或接收的消息的类型。
	RPCMessageTypeKey = attribute.Key("message.type")
	// RPCMessageIDKey 是发送或接收的消息的标识符。
	RPCMessageIDKey = attribute.Key("message.id")
	// RPCMessageCompressedSizeKey 是传输或接收的消息的压缩大小（以字节为单位）。
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")
	// RPCMessageUncompressedSizeKey 是消息的未压缩大小 [以字节为单位发送或接收]
	RPCMessageUncompressedSizeKey = attribute.Key("message.uncompressed_size")
	// ServerEnvironment ...
	ServerEnvironment = attribute.Key("environment")
)

// 常见RPC属性的语义约定
var (
	// RPCSystemGRPC 是 gRPC 作为远程处理系统的语义约定。
	RPCSystemGRPC = semconv.RPCSystemKey.String("grpc")
	// RPCNameMessage 是名为 message 的消息的语义约定。
	RPCNameMessage = RPCNameKey.String("message")
	// RPCMessageTypeSent 是发送的 RPC 消息类型的语义约定。
	RPCMessageTypeSent = RPCMessageTypeKey.String("SENT")
	// RPCMessageTypeReceived 是接收到的 RPC 消息类型的语义约定。
	RPCMessageTypeReceived = RPCMessageTypeKey.String("RECEIVED")
)

// StatusCode Attr 返回一个表示给定 c 的 attribute.Key Value。
func StatusCodeAttr(c gcodes.Code) attribute.KeyValue {
	return GRPCStatusCodeKey.Int64(int64(c))
}
