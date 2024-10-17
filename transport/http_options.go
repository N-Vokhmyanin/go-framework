package transport

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
)

type httpOptions struct {
	matchHeaders     []string
	httpHandlers     []customHttpHandler
	httpMiddlewares  []func(http.Handler) http.Handler
	serverMuxOptions []runtime.ServeMuxOption
	registerHandlers []RegisterHttpHandler
}

type HttpOption func(o *httpOptions)

//goland:noinspection GoUnusedExportedFunction
func WithMatchHeaders(headers ...string) HttpOption {
	return func(o *httpOptions) {
		o.matchHeaders = append(o.matchHeaders, headers...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHttpHandler(handler HttpHandler) HttpOption {
	return func(o *httpOptions) {
		httpHandler := customHttpHandler{
			method:  handler.Method(),
			pattern: handler.Pattern(),
			handler: handler.Handler(),
		}
		o.httpHandlers = append(o.httpHandlers, httpHandler)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHttpHandlerFunc(method string, pattern string, handler HttpHandlerFunc) HttpOption {
	return func(o *httpOptions) {
		httpHandler := customHttpHandler{
			method:  method,
			pattern: pattern,
			handler: handler,
		}
		o.httpHandlers = append(o.httpHandlers, httpHandler)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHttpMiddlewares(middlewares ...func(http.Handler) http.Handler) HttpOption {
	return func(o *httpOptions) {
		o.httpMiddlewares = append(o.httpMiddlewares, middlewares...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithHttpPrependMiddlewares(middlewares ...func(http.Handler) http.Handler) HttpOption {
	return func(o *httpOptions) {
		o.httpMiddlewares = append(middlewares, o.httpMiddlewares...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithServerMuxOptions(options ...runtime.ServeMuxOption) HttpOption {
	return func(o *httpOptions) {
		o.serverMuxOptions = append(o.serverMuxOptions, options...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func WithRegisterHandlers(handlers ...RegisterHttpHandler) HttpOption {
	return func(o *httpOptions) {
		o.registerHandlers = append(o.registerHandlers, handlers...)
	}
}
