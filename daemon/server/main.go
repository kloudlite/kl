package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

const (
	AppPort = 55678
)

// TODO: transform this to grpc for better performance and security

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

	wc := wg_vpn.NewWgClient()

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
			if err := wc.StopService(constants.InterfaceName, true); err != nil {
				fn.Debug(err)
				ch <- err
			}

			w.WriteHeader(http.StatusOK)
			ch <- nil
			return

		case "set-dns":
			var body SetDnsBody
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if body.Dns == "" {
				if err := wc.SetDnsServers([]net.IP{}, constants.InterfaceName, true); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			}

			ips := make([]net.IP, 0)
			for _, v := range strings.Split(body.Dns, ",") {
				ip := net.ParseIP(v)
				if ip == nil {
					http.Error(w, fn.Errorf("invalid ip address: %s", v).Error(), http.StatusInternalServerError)
					return
				}
				ips = append(ips, ip)
			}

			if err := wc.SetDnsServers(ips, constants.InterfaceName, true); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		case "set-search-domain":
			var body SetDomainBody
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if body.Domain == "" {
				if err := wc.ResetSearchDomain(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			}

			if err := wc.SetSearchDomain(body.Domain); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

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
		Addr:    fmt.Sprintf(":%d", AppPort),
		Handler: app,
	}

	fn.Logf("starting server at :%d\n", AppPort)
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
