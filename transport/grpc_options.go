package transport

import (
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
)

type grpcOptions struct {
	prependUnaryInterceptors []grpc.UnaryServerInterceptor
	appendUnaryInterceptors  []grpc.UnaryServerInterceptor

	prependStreamInterceptors []grpc.StreamServerInterceptor
	appendStreamInterceptors  []grpc.StreamServerInterceptor

	serverOptions   []grpc.ServerOption
	recoverOptions  []grpcRecovery.Option
	registerServers []RegisterGrpcServer
	statusMapping   GrpcStatusMappingFunc
}

func newGrpcOptions() *grpcOptions {
	return &grpcOptions{
		statusMapping: DefaultGrpcStatusMapping,
	}
}

type GrpcOption func(o *grpcOptions)

//goland:noinspection GoUnusedExportedFunction
func WithRegisterServers(registers ...RegisterGrpcServer) GrpcOption {
	return func(o *grpcOptions) {
		o.registerServers = append(o.registerServers, registers...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) GrpcOption {
	return func(o *grpcOptions) {
		o.prependUnaryInterceptors = append(o.prependUnaryInterceptors, interceptors...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) GrpcOption {
	return func(o *grpcOptions) {
		o.prependStreamInterceptors = append(o.prependStreamInterceptors, interceptors...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithPrependUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) GrpcOption {
	return func(o *grpcOptions) {
		o.prependUnaryInterceptors = append(interceptors, o.prependUnaryInterceptors...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithPrependStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) GrpcOption {
	return func(o *grpcOptions) {
		o.prependStreamInterceptors = append(interceptors, o.prependStreamInterceptors...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithAppendUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) GrpcOption {
	return func(o *grpcOptions) {
		o.appendUnaryInterceptors = append(o.appendUnaryInterceptors, interceptors...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithAppendStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) GrpcOption {
	return func(o *grpcOptions) {
		o.appendStreamInterceptors = append(o.appendStreamInterceptors, interceptors...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithGrpcServerOptions(opts ...grpc.ServerOption) GrpcOption {
	return func(o *grpcOptions) {
		o.serverOptions = append(o.serverOptions, opts...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithGrpcRecoverOptions(opts ...grpcRecovery.Option) GrpcOption {
	return func(o *grpcOptions) {
		o.recoverOptions = append(o.recoverOptions, opts...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithGrpcStatusMapping(m GrpcStatusMappingFunc) GrpcOption {
	return func(o *grpcOptions) {
		o.statusMapping = m
	}
}
