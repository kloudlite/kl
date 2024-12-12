package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type SetDomainBody struct {
	Domain string `json:"domain"`
}

type SetDnsBody struct {
	Dns string `json:"dns"`
}

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

		case "set-dns":
			var body SetDnsBody
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println("needs to set dns ", body.Dns)

		case "set-search-domain":
			var body SetDomainBody
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println("needs to set search domain ", body.Domain)

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
		Addr:    fmt.Sprintf(":%d", daemon_server.AppPort),
		Handler: app,
	}

	fn.Logf("starting server at :%d", daemon_server.AppPort)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			ch <- err
		}
	}()

	err := <-ch

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	return err
}
