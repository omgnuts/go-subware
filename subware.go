// Copyright 2016 Peanuts. All rights reserved. MIT license.

package subware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Path(r *httprouter.Router, method string, path string) *subware {
	sr := &subware{}
	r.Handle(method, path, sr.serve)
	return sr
}

// SubRouter returns the router generated for the sub route
func (sw *subware) SubRouter() *httprouter.Router {
	r := httprouter.New()
	sw.Use(r).middleware = build(sw.handles)
	sw.locked = true
	return r
}

// Subware is a stack of Middleware Handlers that can be invoked as an http.Handler.
// The middleware stack is run in the sequence that they are added to the stack.
type subware struct {
	middleware middleware
	handles    []mwFunc
	locked     bool
}

// The next http.HandlerFunc is automatically called after the Handler is executed.
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
func (sw *subware) serve(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sw.middleware.serve(w, r, ps)
}

// UseHandler adds a http.Handler onto the middleware stack.
func (sw *subware) Use(handler http.Handler) *subware {
	return sw.UseMWFunc(wrap(handler))
}

// UseHandlerFunc adds a http.HandlerFunc onto the middleware stack.
func (sw *subware) UseFunc(handlerFunc http.HandlerFunc) *subware {
	return sw.UseMWFunc(wrapFunc(handlerFunc))
}

// Use adds a Handle onto the middleware stack.
func (sw *subware) UseHandle(handle httprouter.Handle) *subware {
	return sw.UseMWFunc(wrapHandle(handle))
}

// UseFunc adds a mwFunc function onto the middleware stack.
func (sw *subware) UseMWFunc(fn mwFunc) *subware {
	if !sw.locked {
		sw.handles = append(sw.handles, fn)
	} else {
		panic("Middleware stack must be added before SubRouter() is called")
	}
	return sw
}

// Wrap converts a http.Handler into a HandlerFunc
func wrap(handler http.Handler) mwFunc {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		handler.ServeHTTP(w, r)
		next(w, r, ps)
	}
}

// wrapFunc converts a http.HandlerFunc into a HandlerFunc.
func wrapFunc(fn http.HandlerFunc) mwFunc {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		fn(w, r)
		next(w, r, ps)
	}
}

// wrapHandle converts a httprouter.Handle into a .HandlerFunc.
func wrapHandle(handle httprouter.Handle) mwFunc {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		handle(w, r, ps)
		next(w, r, ps)
	}
}

// The stack is traversed using a linked-list handler interface that provides
// every middleware a forward reference to the next middleware in the stack.
type mwFunc func(http.ResponseWriter, *http.Request, httprouter.Params, httprouter.Handle)

// Each Middleware should yield to the next middleware in the chain by invoking the next http.HandlerFunc
type middleware struct {
	fn   mwFunc
	next *middleware
}

// The next http.HandlerFunc is automatically called after the Handler is executed.
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
func (m middleware) serve(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	m.fn(w, r, ps, m.next.serve)
}

func build(fns []mwFunc) middleware {
	var next middleware

	if len(fns) == 0 {
		return empty()
	} else if len(fns) > 1 {
		next = build(fns[1:])
	} else {
		next = empty()
	}

	return middleware{fns[0], &next}
}

func empty() middleware {
	return middleware{
		func(http.ResponseWriter, *http.Request, httprouter.Params, httprouter.Handle) { /* do nothing */ },
		&middleware{},
	}
}
