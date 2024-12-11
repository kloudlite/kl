package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Server struct {
	bin string
}

func New(binName string) *Server {
	return &Server{
		bin: binName,
	}
}
func portAvailable(port string) bool {
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

type StreamingWriter struct {
	io.Writer
}

func (w StreamingWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return n, err
}

func (s *Server) Start(ctx context.Context) error {

	ch := make(chan error)

	defer ctx.Done()

	app := http.NewServeMux()
	app.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// outputCh := make(chan string)
		errCh := make(chan error)

		command := strings.TrimPrefix(req.URL.Path, "/")

		switch command {
		case "healthy":
			w.WriteHeader(http.StatusOK)
			return

		case "exit":
			w.WriteHeader(http.StatusOK)
			ch <- nil
			return

		case "start", "stop", "status", "restart":
			if err := fn.StreamOutput(req.Context(), fmt.Sprintf("%s vpn %s", s.bin, command), map[string]string{"KL_APP": "true"}, StreamingWriter{Writer: w}, errCh); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", proxy.AppPort),
		Handler: app,
	}

	fn.Logf("starting server at :%d", proxy.AppPort)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			ch <- err
		}
	}()

	err := <-ch

	if err2 := server.Shutdown(ctx); err2 != nil {
		return err2
	}

	return err
}
