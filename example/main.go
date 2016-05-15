// Copyright 2016 Peanuts. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	jr "github.com/omgnuts/jrouter"
)

func main() {
	// Create a new router. The API is the same as httprouter.New()
	router := httprouter.New()
	router.GET("/public/post/:id", appHandler("viewing: /public/post/:id"))
	router.GET("/inlinefunc", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		w.Write([]byte("Hello from an inline func!"))
	})

	// Create a subrouter using mainRouter.Path(method, path)
	// Add in the required middleware
	pttRouter := jr.Path(router, "GET", "/protected/*path").
		UseFunc(middlewareA).
		UseHandle(middlewareB).
		UseMWFunc(middlewareC).
		SubRouter()
	{
		pttRouter.GET("/protected/user/:id", appHandler("viewing: /protected/user/:id"))
		pttRouter.GET("/protected/users", appHandler("viewing: /protected/users"))
	}

	// Another way to fire up a subroute is as follows.
	subware := jr.Path(router, "GET", "/admin/*path")
	subware.UseMWFunc(middlewareC)
	admRouter := subware.SubRouter()
	{
		admRouter.GET("/admin/log/:id", appHandler("viewing: /admin/log/:id"))
		admRouter.GET("/admin/stats", appHandler("viewing: /admin/stats"))
	}

	// Start the server with the main router
	log.Fatal(http.ListenAndServe(":8080", router))
}

func appHandler(msg string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		if id != "" {
			w.Write([]byte("[PARAM] id = " + id + "\n"))
		}
		w.Write([]byte("[OUTPUT] " + msg + "\n"))
	}
}

func middlewareA(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("[jr] I am middlewareA \n"))
}

func middlewareB(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("[jr] I am middlewareB \n"))
}

func middlewareC(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
	w.Write([]byte("[jr] I am middlewareC \n"))
	next(w, r, ps)
}
