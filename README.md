# Jrouter [![Build Status](https://travis-ci.org/omgnuts/jrouter.svg?branch=master)](https://travis-ci.org/omgnuts/jrouter) [![Coverage](http://gocover.io/_badge/github.com/omgnuts/jrouter?0)](http://gocover.io/github.com/omgnuts/jrouter) [![GoDoc](https://godoc.org/github.com/omgnuts/jrouter?status.svg)](http://godoc.org/github.com/omgnuts/jrouter)

Jrouter shows how you can extend Httprouter in a simple, non-intrusive manner to make it easier to support
subroutes and sublevel middleware. This allows you to easily apply different sub-level middleware that may be
specific only to certain subgroup routes. For example to apply authentication middleware at various subroutes.
The key objective in this extension is to preserve the lightweight beauty and high performance of httprouter.

### SubRouting Example

Here's a basic example of how subrouting can be done with jrouter. The examples are provided in the source.

```go
import (
    "log"
    "net/http"
    jr "github.com/omgnuts/jrouter"
)

func main() {
    router := httprouter.New()
    router.GET("/", index)

    subrouter := jr.Path(router, "GET", "/protected/*path").
        UseFunc(middlewareA).
        UseHandle(middlewareB).
        UseMWFunc(middlewareC).
        SubRouter()
    {
        subrouter.GET("/protected/user/:id", appHandler("viewing: /protected/user/:id"))
        subrouter.GET("/protected/users", appHandler("viewing: /protected/users"))
    }

    log.Fatal(http.ListenAndServe(":8080", router))
}

// Below are sample handlers with various method signatures

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "Welcome!\n")
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
```
