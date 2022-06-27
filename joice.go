package joice

import (
	"net/http"
)

// New returns a new Joice instance with no middleware preconfigured.
func NewMW() *Joice {
	return &Joice{}
}

// Joice is a stack of Middleware Handlers that can be invoked as an http.Handler.
// The middleware stack is run in the sequence that they are added to the stack.
type Joice struct {
	middleware middleware
	handles   []Func
}

// Joice itself is a http.Handler. This allows it to used as a substack manager
func (n *Joice) ServeHTTP(w http.ResponseWriter, r *http.Request, ps Params) {
	n.middleware.serve(w, r, ps)
}

// UseHandler adds a http.Handler onto the middleware stack.
func (n *Joice) Use(handler http.Handler) *Joice {
	return n.UseHandlerFunc(wrap(handler))
}

// UseHandlerFunc adds a http.HandlerFunc onto the middleware stack.
func (n *Joice) UseFunc(handlerFunc http.HandlerFunc) *Joice {
	return n.UseHandlerFunc(wrapFunc(handlerFunc))
}

// Use adds a joice.Handler onto the middleware stack.
func (n *Joice) UseHandler(handler Handler) *Joice {
	return n.UseHandlerFunc(handler.ServeHTTP)
}

// UseFunc adds a joice.Func function onto the middleware stack.
func (n *Joice) UseHandlerFunc(fn Func) *Joice {
	n.handles = append(n.handles, fn)
	n.middleware = build(n.handles)
	return n
}

// Use adds a joice.Handler onto the middleware stack.
func (n *Joice) UseParamHandle(handle Handle) *Joice {
	return n.UseHandlerFunc(wrapParamHandle(handle))
}

// Handler exposes an adapter to support specific middleware that uses this signature
// ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, Params, Handle)
}

// joice moves the stack by using a linked-list handler interface that provides
// every middleware a forward reference to the next middleware in the stack.
type Func func(http.ResponseWriter, *http.Request, Params, Handle)

// Each Middleware should yield to the next middleware in the chain by invoking the next http.HandlerFunc
type middleware struct {
	fn 		Func
	next  *middleware
}

// The next http.HandlerFunc is automatically called after the Handler is executed.
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
func (m middleware) serve(w http.ResponseWriter, r *http.Request, ps Params) {
	m.fn(w, r, ps, m.next.serve)
}

// Wrap converts a http.Handler into a joice.HandlerFunc
func wrap(handler http.Handler) Func {
	return func(w http.ResponseWriter, r *http.Request, ps Params, next Handle) {
		handler.ServeHTTP(w, r)
		next(w, r, ps)
	}
}

// wrapFunc converts a http.HandlerFunc into a joice.HandlerFunc.
func wrapFunc(fn http.HandlerFunc) Func {
	return func(w http.ResponseWriter, r *http.Request, ps Params, next Handle) {
		fn(w, r)
		next(w, r, ps)
	}
}

// wrapParamFunc converts a httprouter.Handle into a joice.HandlerFunc.
func wrapParamHandle(handle Handle) Func {
	return func(w http.ResponseWriter, r *http.Request, ps Params, next Handle) {
		handle(w, r, ps)
		next(w, r, ps)
	}
}

func build(handles []Func) middleware {
	var next middleware

	if len(handles) == 0 {
		return empty()
	} else if len(handles) > 1 {
		next = build(handles[1:])
	} else {
		next = empty()
	}

	return middleware{ handles[0], &next }
}

func empty() middleware {
	return middleware{
		func(http.ResponseWriter, *http.Request, Params, Handle) { /* do nothing */ },
		&middleware{},
	}
}