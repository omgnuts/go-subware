// Copyright 2016 Peanuts. All rights reserved. MIT license.

package subware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
)

/* Test Helpers */

// Void handler to do nothing
func voidHandler(http.ResponseWriter, *http.Request, httprouter.Params) {}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func assertTrue(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func assertFalse(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func assertNoPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

// Ensures that subware can chain link its handlers.
func TestSubwareServe(t *testing.T) {
	result := ""
	rec := httptest.NewRecorder()

	ns := subware{}
	ns.UseMWFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		result += "Live "
		next(w, r, ps)
		result += "fullest!"
	})
	ns.UseMWFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		result += "Life "
		next(w, r, ps)
		result += "the "
	})
	ns.UseMWFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		result += "to "
		w.WriteHeader(http.StatusBadRequest)
	})

	ns.middleware = build(ns.handles)
	ns.serve(rec, (*http.Request)(nil), nil)

	assertTrue(t, result, "Live Life to the fullest!")
	assertTrue(t, rec.Code, http.StatusBadRequest)
}

func TestSubwareUseFunc(t *testing.T) {
	result := ""
	response := httptest.NewRecorder()

	n1 := subware{}
	n1.UseFunc(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		result = "one"
	}))
	n1.UseFunc(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		result += "two"
	}))

	n1.middleware = build(n1.handles)
	n1.serve(response, (*http.Request)(nil), nil)
	assertTrue(t, 2, len(n1.handles))
	assertTrue(t, result, "onetwo")

	n2 := n1.With(func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		result += "three"
	})

	// Verify that n1 was not affected and unchanged.
	n1.serve(response, (*http.Request)(nil), nil)
	assertTrue(t, 2, len(n1.handles))
	assertTrue(t, result, "onetwo")

	n2.middleware = build(n2.handles)
	n2.serve(response, (*http.Request)(nil), nil)
	assertTrue(t, 3, len(n2.handles))
	assertTrue(t, result, "onetwothree")
}

// Ensures that the middleware chain
// can correctly return all of its handlers.
func TestSubwareUseMWFunc(t *testing.T) {
	rec := httptest.NewRecorder()
	ns := subware{}
	handles := ns.handles
	assertTrue(t, 0, len(handles))

	ns.UseMWFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
		w.WriteHeader(http.StatusOK)
	})

	// Expects the length of handlers to be exactly 1
	// after adding exactly one handler to the middleware chain
	handles = ns.handles
	assertTrue(t, 1, len(handles))

	// Ensures that the first handler that is in sequence behaves
	// exactly the same as the one that was registered earlier
	handles[0](rec, (*http.Request)(nil), nil, nil)
	assertTrue(t, rec.Code, http.StatusOK)
}

// Ensures that the middleware chain
// can correctly return all of its handlers.
func TestSubwareUseHandle(t *testing.T) {
	rec := httptest.NewRecorder()
	ns := subware{}
	handles := ns.handles
	assertTrue(t, 0, len(handles))

	ns.UseHandle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})

	// Expects the length of handlers to be exactly 1
	// after adding exactly one handler to the middleware chain
	handles = ns.handles
	assertTrue(t, 1, len(handles))

	// Ensures that the first handler that is in sequence behaves
	// exactly the same as the one that was registered earlier
	handles[0](rec, (*http.Request)(nil), nil, voidHandler)
	assertTrue(t, rec.Code, http.StatusOK)
}

// Ensure that the Subware can handle nil handles
func TestSubwareUseNil(t *testing.T) {
	assertPanic(t, func() {
		ns := subware{}
		ns.Use(nil)
	})

	assertPanic(t, func() {
		ns := subware{}
		ns.UseFunc(nil)
	})

	assertPanic(t, func() {
		ns := subware{}
		ns.UseHandle(nil)
	})
}

// Test for function wrap
func TestWrap(t *testing.T) {
	response := httptest.NewRecorder()

	handler := wrap(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	handler(response, (*http.Request)(nil), nil, voidHandler)

	assertTrue(t, response.Code, http.StatusOK)
}

// Test for function wrapFunc
func TestWrapFunc(t *testing.T) {
	response := httptest.NewRecorder()

	// WrapFunc(f) equals Wrap(http.HandlerFunc(f)), it's simpler and useful.
	handler := wrapFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler(response, (*http.Request)(nil), nil, voidHandler)
	assertTrue(t, response.Code, http.StatusOK)
}

// Test for function wrapHandle
func TestWrapHandle(t *testing.T) {
	response := httptest.NewRecorder()

	// WrapFunc(f) equals Wrap(http.HandlerFunc(f)), it's simpler and useful.
	handler := wrapHandle(func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		rw.WriteHeader(http.StatusOK)
	})

	handler(response, (*http.Request)(nil), nil, voidHandler)

	assertTrue(t, response.Code, http.StatusOK)
}

// Ensure that the Subware can handle nil handles
func TestSubwareNoHandles(t *testing.T) {
	assertNoPanic(t, func() {
		ns := subware{}
		ns.middleware = build(ns.handles)
	})
}

func TestNoUseAfterLocked(t *testing.T) {
	assertPanic(t, func() {
		ns := subware{}
		ns.SubRouter()
		ns.UseMWFunc(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
			next(w, r, ps)
		})
	})
}

func TestSubwareRun(t *testing.T) {
	router := httprouter.New()
	r1 := Path(router, "GET", "/").SubRouter()
	// just test that Run doesn't bomb
	go http.ListenAndServe(":8080", r1)
}
