package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
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

	fs := NewSimpleServer(dir)
	log.Println(fs.ListenAndServe(*addr))
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func tabbedInterfaces() ([]byte, error) {

	buf := bytes.Buffer{}
	w := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "Name\tAddresses\tFlags\t\n")

	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, a := range ifs {
		addrs := []string{}
		ifAddrs, err := a.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range ifAddrs {
			addrs = append(addrs, addr.String())
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", a.Name, strings.Join(addrs, ", "), a.Flags.String())

	}
	w.Flush()
	bytes, err := ioutil.ReadAll(&buf)
	return bytes, err

}

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
// If srv.Addr is blank, ":http" is used.
// ListenAndServe always returns a non-nil error.
func (s *SimpleServer) ListenAndServe(addr string) error {
	server := &http.Server{Addr: addr, Handler: s.handler}
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Serving folder %s at %s", dir, ln.Addr())

	bytes, err := tabbedInterfaces()
	if err != nil {
		return err
	}
	fmt.Println("Interface Addresses:")
	fmt.Println(string(bytes))

	return server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
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
