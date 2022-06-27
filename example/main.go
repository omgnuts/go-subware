package main

import (
	"net/http"
	"fmt"
	"github.com/omgnuts/joice"
	"github.com/jaem/nimble"
)

func main() {
	mux := joice.New()
	mux.GET("/hello/*watch", flush("Hello!"))
	mux.GET("/helloinline", func(w http.ResponseWriter, req *http.Request, _ joice.Params) {
		fmt.Fprintf(w, "Hello inline!")
	})

	auth := joice.New()
	{
		auth.GET("/auth/boy/:watch", flush("boy"))
		auth.GET("/auth/girl", flush("girl"))
	}

	sub := joice.NewMW()
	sub.UseParamHandle(middlewareA)
	sub.UseParamHandle(middlewareB)
	sub.Use(auth)

	mux.GET("/auth/*sub", sub.ServeHTTP)

	n := nimble.Default()
	n.Use(mux)
	n.Run(":3000")
}

func flush(msg string) joice.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps joice.Params) {
		fmt.Println("...." + ps.ByName("watch"))
		fmt.Fprintf(w, msg)
	}
}

func middlewareA(w http.ResponseWriter, r *http.Request, _ joice.Params) {
	fmt.Println("[nim.] I am middlewareA")
	//bun := hax.GetBundle(c)
	//bun.Set("valueA", ": from middlewareA")
}

func middlewareB(w http.ResponseWriter, r *http.Request, _ joice.Params) {
	fmt.Println("[nim.] I am middlewareB")
	//bun := hax.GetBundle(c)
	//bun.Set("valueB", ": from middlewareB")
}