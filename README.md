# Subware [![Build Status](https://travis-ci.org/omgnuts/subware.svg?branch=master)](https://travis-ci.org/omgnuts/subware)  [![GoDoc](https://godoc.org/github.com/omgnuts/subware?status.svg)](http://godoc.org/github.com/omgnuts/subware)

Subware shows how you can extend [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
in a simple, non-intrusive manner to support subroutes and sub-level grouped middleware.

This allows you to easily apply different sub-level middleware that may be
specific only to certain subgroup routes. For example to apply authentication middleware at various subroutes.
The key purpose in this extension is to preserve the lightweight beauty and high performance of httprouter.

Subware uses httprouter, but you can probably modify it for other lightweight routers as well.

Hope this helps! ;)

### Quick example

Here's a basic example of how subrouting can be done with subware. The examples are provided in the source.

```go
import (
    "log"
    "net/http"
    "github.com/omgnuts/subware"
)

func main() {
    router := httprouter.New()
    router.GET("/", index)

    subrouter := subware.Path(router, "GET", "/protected/*path").
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
	w.Write([]byte("[sw] I am middlewareA \n"))
}

func middlewareB(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("[sw] I am middlewareB \n"))
}

func middlewareC(w http.ResponseWriter, r *http.Request, ps httprouter.Params, next httprouter.Handle) {
	w.Write([]byte("[sw] I am middlewareC \n"))
	next(w, r, ps)
}
```

### Full code example

Here's the code to run the examples:

```go
# go run example/main.go
```

Then visit the following links on your browser.

```
http://localhost:8080/inlinefunc
http://localhost:8080/public/post/12345

http://localhost:8080/protected/user/batman
http://localhost:8080/protected/users

http://localhost:8080/admin/log/54321
http://localhost:8080/admin/stats
```

_**Final Note: Subware provides convenience methods to perform subrouting with htttprouter.
This is not a router by itself, there are a great many out there.**_

That's all folks!

---

MIT License
