package core

import (
	"io"
	"net"
	"net/http"

	"golang.org/x/net/proxy"
)

// HttpProxyRoutineHandler .....
//
//
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
	httpConn, _, err := hijack.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodConnect {
		httpConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	} else {
		r.Write(socksConn)
	}

	pipeConn := func(w, r net.Conn) {
		io.Copy(w, r)
		if c, ok := w.(*net.TCPConn); ok {
			c.CloseWrite()
		}
		if c, ok := r.(*net.TCPConn); ok {
			c.CloseRead()
		}
	}
	go pipeConn(socksConn, httpConn)
	go pipeConn(httpConn, socksConn)
}
