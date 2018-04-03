package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

var (
	addr = flag.String("addr", ":8080", "address")
	dirF = flag.String("dir", ".", "dir to serve")
	dir  string
)

func main() {
	flag.Parse()
	dir, err := filepath.Abs(*dirF)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Serving folder %s at %s", dir, *addr)

	server := NewSimpleServer(dir)

	fmt.Println(http.ListenAndServe(*addr, server))
}

// SimpleServer wraps a fileserver and adds some request logging
type SimpleServer struct {
	Dir     string
	handler http.Handler
}

// NewSimpleServer creates a new SimpleServer
func NewSimpleServer(dir string) *SimpleServer {
	return &SimpleServer{
		Dir:     dir,
		handler: http.FileServer(http.Dir(dir)),
	}
}

// ServeHTTP implements http.Handler
func (s *SimpleServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s requested file %s\n", r.RemoteAddr, r.URL.Path)
	go func() {
		<-r.Context().Done()
		log.Printf("%s request file %s completed", r.RemoteAddr, r.URL.Path)
	}()
	s.handler.ServeHTTP(w, r)
}
