package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"sync"

	"golang.org/x/net/proxy"
)

type HttpProxyRoutineHandler struct {
	Dialer proxy.Dialer
}

func (h *HttpProxyRoutineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hijack, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	port := r.URL.Port()
	if port == "" {
		port = "80"
	}
	socksConn, err := h.Dialer.Dial("tcp", r.URL.Hostname()+":"+port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	defer socksConn.Close()
	httpConn, _, err := hijack.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer httpConn.Close()
	if r.Method == http.MethodConnect {
		httpConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	} else {
		r.Write(socksConn)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go transfer(httpConn, socksConn, &wg)
	go transfer(socksConn, httpConn, &wg)
	wg.Wait()

}

func transfer(src, dst net.Conn, wg *sync.WaitGroup) {
	io.Copy(dst, src)
	wg.Done()
}

func main() {
	httpAddr := flag.String("http", "0.0.0.0:8000", "local http proxy address")
	socks5Addr := flag.String("socks5", "socks5://127.0.0.1:1080", "remote socks5 address")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	socksURL, err := url.Parse(*socks5Addr)
	if err != nil {
		log.Fatalln("proxy url parse error:", err)
	}
	socks5Dialer, err := proxy.FromURL(socksURL, proxy.Direct)
	if err != nil {
		log.Fatalln("can not make proxy dialer:", err)
	}
	if err := http.ListenAndServe(*httpAddr, &HttpProxyRoutineHandler{Dialer: socks5Dialer}); err != nil {
		log.Fatalln("can not start http server:", err)
	}

}
