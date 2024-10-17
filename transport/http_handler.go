package transport

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
)

type HttpHandlerFunc func(w http.ResponseWriter, r *http.Request, params map[string]string) error

type HttpHandler interface {
	Method() string
	Pattern() string
	Handler() HttpHandlerFunc
}

type customHttpHandler struct {
	method  string
	pattern string
	handler HttpHandlerFunc
}

var _ HttpHandler = (*customHttpHandler)(nil)

func (h customHttpHandler) Method() string {
	return h.method
}

func (h customHttpHandler) Pattern() string {
	return h.pattern
}

func (h customHttpHandler) Handler() HttpHandlerFunc {
	return h.handler
}

func (h customHttpHandler) HandlerFunc(mux *runtime.ServeMux) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(r.Context())
		req := r.WithContext(ctx)
		defer cancel()

		if err := h.handler(w, req, pathParams); err != nil {
			_, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
		}
	}
}
