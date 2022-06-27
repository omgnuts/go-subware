# Jrouter [![Build Status](https://travis-ci.org/omgnuts/jrouter.svg?branch=joice)](https://travis-ci.org/omgnuts/jrouter)  [![GoDoc](https://godoc.org/github.com/omgnuts/jrouter?status.svg)](http://godoc.org/github.com/omgnuts/jrouter)

Jrouter helps you to extend Httprouter, in an unobstrusive manner to make it easier to support
subroutes and sublevel middleware. This allows you to easily apply different sub-level middleware that may be
specific only to certain subgroup routes. For example to apply authentication middleware at various subroutes.
The key objective in this extension is to preserve the lightweight beauty and high performance of httprouter.
Core changes are in jrouter.go

In contrast to the [default mux][http.ServeMux] of Go's `net/http` package, this router supports variables
in the routing pattern and matches against the request method. It also scales better. The router is optimized
for high performance and a small memory footprint. It scales well even with very
long paths and a large number of routes. A compressing dynamic trie (radix tree) structure is used for
efficient matching.

### SubRouting Example

Here's a basic example of how subrouting can be done with jrouter. The examples are provided in the source.

~~~go
import (
    "log"
    "net/http"
    jr "github.com/omgnuts/jrouter"
)

func main() {
    router := jr.New()
    router.GET("/", index)

    subrouter := router.Path("GET", "/protected/*path").
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

func index(w http.ResponseWriter, r *http.Request, _ jrouter.Params) {
    fmt.Fprint(w, "Welcome!\n")
}

func appHandler(msg string) jr.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps jr.Params) {
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

func middlewareB(w http.ResponseWriter, r *http.Request, ps jr.Params) {
	w.Write([]byte("[jr] I am middlewareB \n"))
}

func middlewareC(w http.ResponseWriter, r *http.Request, ps jr.Params, next jr.Handle) {
	w.Write([]byte("[jr] I am middlewareC \n"))
	next(w, r, ps)
}
~~~

## Inherited Features (from HttpRouter)

**Only explicit matches:** With other routers, like [`http.ServeMux`][http.ServeMux], a requested URL path
 could match multiple patterns. Therefore they have some awkward pattern priority rules, like *longest match*
 or *first registered, first matched*. By design of this router, a request can only match exactly one or no route.
 As a result, there are also no unintended matches, which makes it great for SEO and improves the user experience.

**Stop caring about trailing slashes:** Choose the URL style you like, the router automatically redirects
the client if a trailing slash is missing or if there is one extra. Of course it only does so, if the new path
has a handler. If you don't like it, you can [turn off this behavior](https://godoc.org/github.com/omgnuts/jrouter#Router.RedirectTrailingSlash).

**Path auto-correction:** Besides detecting the missing or additional trailing slash at no extra cost,
the router can also fix wrong cases and remove superfluous path elements (like `../` or `//`).
Is [CAPTAIN CAPS LOCK](http://www.urbandictionary.com/define.php?term=Captain+Caps+Lock) one of your users?
Jrouter can help him by making a case-insensitive look-up and redirecting him to the correct URL.

**Parameters in your routing pattern:** Stop parsing the requested URL path, just give the path segment a name
and the router delivers the dynamic value to you. Because of the design of the router, path parameters are very cheap.

**Zero Garbage:** The matching and dispatching process generates zero bytes of garbage. In fact,
the only heap allocations that are made, is by building the slice of the key-value pairs for path parameters.
If the request path contains no parameters, not a single heap allocation is necessary.

**Best Performance:** [Benchmarks speak for themselves][benchmark].
See below for technical details of the implementation.

**No more server crashes:** You can set a [Panic handler][Router.PanicHandler] to deal with panics
occurring during handling a HTTP request. The router then recovers and lets the `PanicHandler` log what
happened and deliver a nice error page.

**Perfect for APIs:** The router design encourages to build sensible, hierarchical RESTful APIs.
Moreover it has builtin native support for [OPTIONS requests](http://zacstewart.com/2012/04/14/http-options-method.html)
and `405 Method Not Allowed` replies.

Of course you can also set **custom [`NotFound`][Router.NotFound]
and [`MethodNotAllowed`](https://godoc.org/github.com/omgnuts/jrouter#Router.MethodNotAllowed)
handlers** and [**serve static files**][Router.ServeFiles].

## Usage

This is just a quick introduction, view the [GoDocs](http://godoc.org/github.com/omgnuts/jrouter) for details.

Let's start with a trivial example:

```go
package main

import (
    "fmt"
    "github.com/omgnuts/jrouter"
    "net/http"
    "log"
)

func Index(w http.ResponseWriter, r *http.Request, _ jrouter.Params) {
    fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps jrouter.Params) {
    fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
    router := jrouter.New()
    router.GET("/", Index)
    router.GET("/hello/:name", Hello)

    log.Fatal(http.ListenAndServe(":8080", router))
}
```

### Named parameters

As you can see, `:name` is a *named parameter*. The values are accessible via `jrouter.Params`,
which is just a slice of `jrouter.Param`. You can get the value of a parameter either by its
index in the slice, or by using the `ByName(name)` method: `:name` can be obtained by `ByName("name")`.

Named parameters only match a single path segment:

```
Pattern: /user/:user

 /user/gordon              match
 /user/you                 match
 /user/gordon/profile      no match
 /user/                    no match
```

**Note:** Since this router has only explicit matches, you can not register static routes
and parameters for the same path segment. For example you can not register the patterns `/user/new`
and `/user/:user` for the same request method at the same time. The routing of different request methods
is independent from each other.

### Catch-All parameters

The second type are *catch-all* parameters and have the form `*name`. Like the name suggests,
they match everything. Therefore they must always be at the **end** of the pattern:

```
Pattern: /src/*filepath

 /src/                     match
 /src/somefile.go          match
 /src/subdir/somefile.go   match
```

## How does it work?

The router relies on a tree structure which makes heavy use of *common prefixes*,
it is basically a *compact* [*prefix tree*](https://en.wikipedia.org/wiki/Trie)
(or just [*Radix tree*](https://en.wikipedia.org/wiki/Radix_tree)). Nodes with a common prefix
also share a common parent. Here is a short example what the routing tree for the `GET` request method
could look like:

```
Priority   Path             Handle
9          \                *<1>
3          ├s               nil
2          |├earch\         *<2>
1          |└upport\        *<3>
2          ├blog\           *<4>
1          |    └:post      nil
1          |         └\     *<5>
2          ├about-us\       *<6>
1          |        └team\  *<7>
1          └contact\        *<8>
```

Every `*<num>` represents the memory address of a handler function (a pointer). If you follow a path
through the tree from the root to the leaf, you get the complete route path, e.g `\blog\:post\`,
where `:post` is just a placeholder ([*parameter*](#named-parameters)) for an actual post name.
Unlike hash-maps, a tree structure also allows us to use dynamic parts like the `:post` parameter,
since we actually match against the routing patterns instead of just comparing hashes.
[As benchmarks show][benchmark], this works very well and efficient.

Since URL paths have a hierarchical structure and make use only of a limited set of characters (byte values),
it is very likely that there are a lot of common prefixes. This allows us to easily reduce the routing into
ever smaller problems. Moreover the router manages a separate tree for every request method. For one thing
it is more space efficient than holding a method->handle map in every single node, for another thing is also
allows us to greatly reduce the routing problem before even starting the look-up in the prefix-tree.

For even better scalability, the child nodes on each tree level are ordered by priority, where the
priority is just the number of handles registered in sub nodes (children, grandchildren, and so on..).
This helps in two ways:

1. Nodes which are part of the most routing paths are evaluated first. This helps to make as much routes
as possible to be reachable as fast as possible.
2. It is some sort of cost compensation. The longest reachable path (highest cost) can always be evaluated
first. The following scheme visualizes the tree structure. Nodes are evaluated from top to bottom
and from left to right.

```
├------------
├---------
├-----
├----
├--
├--
└-
```

## Why doesn't this work with `http.Handler`?

**It does!** The router itself implements the `http.Handler` interface. Moreover the router provides
convenient [adapters for `http.Handler`][Router.Handler]s and [`http.HandlerFunc`][Router.HandlerFunc]s
which allows them to be used as a [`jrouter.Handle`][Router.Handle] when registering a route. The only
disadvantage is, that no parameter values can be retrieved when a `http.Handler` or `http.HandlerFunc` is used,
since there is no efficient way to pass the values with the existing function parameters.
Therefore [`jrouter.Handle`][Router.Handle] has a third function parameter.

Just try it out for yourself, the usage of jrouter is very straightforward. The package is compact and minimalistic,
but also probably one of the easiest routers to set up.

## Where can I find Middleware *X*?

This package just provides a very efficient request router with a few extra features.
The router is just a [`http.Handler`][http.Handler], you can chain any http.Handler
compatible middleware before the router, for example the [Gorilla handlers](http://www.gorillatoolkit.org/pkg/handlers).
Or you could [just write your own](https://justinas.org/writing-http-middleware-in-go/), it's very easy!


### Multi-domain / Sub-domains

Here is a quick example: Does your server serve multiple domains / hosts?
You want to use sub-domains?
Define a router per host!

```go
// We need an object that implements the http.Handler interface.
// Therefore we need a type for which we implement the ServeHTTP method.
// We just use a map here, in which we map host names (with port) to http.Handlers
type HostSwitch map[string]http.Handler

// Implement the ServerHTTP method on our new type
func (hs HostSwitch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if a http.Handler is registered for the given host.
	// If yes, use it to handle the request.
	if handler := hs[r.Host]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		// Handle host names for wich no handler is registered
		http.Error(w, "Forbidden", 403) // Or Redirect?
	}
}

func main() {
	// Initialize a router as usual
	router := jrouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	// Make a new HostSwitch and insert the router (our http handler)
	// for example.com and port 12345
	hs := make(HostSwitch)
	hs["example.com:12345"] = router

	// Use the HostSwitch to listen and serve on port 12345
	log.Fatal(http.ListenAndServe(":12345", hs))
}
```

### Basic Authentication

Another quick example: Basic Authentication (RFC 2617) for handles:

```go
package main

import (
    "bytes"
    "encoding/base64"
    "fmt"
    "github.com/omgnuts/jrouter"
    "net/http"
    "log"
    "strings"
)

func BasicAuth(h jrouter.Handle, user, pass []byte) jrouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps jrouter.Params) {
		const basicAuthPrefix string = "Basic "

		// Get the Basic Authentication credentials
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 &&
					bytes.Equal(pair[0], user) &&
					bytes.Equal(pair[1], pass) {

					// Delegate request to the given handle
					h(w, r, ps)
					return
				}
			}
		}

		// Request Basic Authentication otherwise
		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func Index(w http.ResponseWriter, r *http.Request, _ jrouter.Params) {
    fmt.Fprint(w, "Not protected!\n")
}

func Protected(w http.ResponseWriter, r *http.Request, _ jrouter.Params) {
    fmt.Fprint(w, "Protected!\n")
}

func main() {
    user := []byte("gordon")
    pass := []byte("secret!")
    
    router := jrouter.New()
    router.GET("/", Index)
    router.GET("/protected/", BasicAuth(Protected, user, pass))

    log.Fatal(http.ListenAndServe(":8080", router))
}
```

## Chaining with the NotFound handler

**NOTE: It might be required to set [`Router.HandleMethodNotAllowed`][Router.HandleMethodNotAllowed]
to `false` to avoid problems.**

You can use another [`http.Handler`][http.Handler], for example another router, to handle requests
which could not be matched by this router by using the [`Router.NotFound`][Router.NotFound] handler.
This allows chaining.

### Static files

The `NotFound` handler can for example be used to serve static files from the root path `/`
(like an `index.html` file along with other assets):

```go
// Serve static files from the ./public directory
router.NotFound = http.FileServer(http.Dir("public"))
```

But this approach sidesteps the strict core rules of this router to avoid routing problems.
A cleaner approach is to use a distinct sub-path for serving files, like `/static/*filepath` or `/files/*filepath`.

[benchmark]: <https://github.com/julienschmidt/go-http-routing-benchmark>
[http.Handler]: <https://golang.org/pkg/net/http/#Handler
[http.ServeMux]: <https://golang.org/pkg/net/http/#ServeMux>
[Router.Handle]: <https://godoc.org/github.com/omgnuts/jrouter#Router.Handle>
[Router.HandleMethodNotAllowed]: <https://godoc.org/github.com/omgnuts/jrouter#Router.HandleMethodNotAllowed>
[Router.Handler]: <https://godoc.org/github.com/omgnuts/jrouter#Router.Handler>
[Router.HandlerFunc]: <https://godoc.org/github.com/omgnuts/jrouter#Router.HandlerFunc>
[Router.NotFound]: <https://godoc.org/github.com/omgnuts/jrouter#Router.NotFound>
[Router.PanicHandler]: <https://godoc.org/github.com/omgnuts/jrouter#Router.PanicHandler>
[Router.ServeFiles]: <https://godoc.org/github.com/omgnuts/jrouter#Router.ServeFiles>
