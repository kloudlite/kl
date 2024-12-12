package daemon_server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func GetUserHomeDir() (string, error) {
	if runtime.GOOS == "windows" {
		return xdg.Home, nil
	}

	if euid := os.Geteuid(); euid == 0 {
		username, ok := os.LookupEnv("SUDO_USER")
		if !ok {
			return "", errors.New("failed to get sudo user name")
		}

		oldPwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		sp := strings.Split(oldPwd, "/")

		for i := range sp {
			if sp[i] == username {
				return path.Join("/", path.Join(sp[:i+1]...)), nil
			}
		}

		return "", errors.New("failed to get home path of sudo user")
	}

	userHome, ok := os.LookupEnv("HOME")
	if !ok {
		return "", errors.New("failed to get home path of user")
	}

	return userHome, nil
}

type Proxy struct {
	logResponse bool
}

func NewProxyWithService(logResponse bool, ensureAppRunning ...bool) (*Proxy, error) {

	p := &Proxy{
		logResponse: logResponse,
	}

	if len(ensureAppRunning) > 0 && !ensureAppRunning[0] {
		return p, nil
	}

	if p.Status() {
		return p, nil
	}

	count := 0
	for {
		if p.Status() {
			return p, nil
		}

		if runtime.GOOS != "windows" {
			fn.Log(text.Yellow("Administrator privileges are required to configure the network settings. Please enter your password to proceed."))

			cmd := exec.Command("sudo", "echo", "")
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			err := cmd.Run()
			if err != nil {
				return nil, err
			}
			command := exec.Command("sudo", flags.CliName, "app", "start")
			_ = command.Start()

		} else {
			if _, err := fn.WinSudoExec(fmt.Sprintf("%s app start", flags.CliName), nil); err != nil {
				functions.PrintError(err)
			}
		}

		count++
		if count >= 2 {
			return nil, fn.Errorf("failed to start app")
		}

		time.Sleep(2 * time.Second)
	}

}

func (p *Proxy) MakeRequest(path string, params ...[]byte) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%d%s", AppPort, path)

	if err := func() error {
		hostIp := "localhost"

		url = fmt.Sprintf("http://%s:%d%s", hostIp, AppPort, path)
		return nil
	}(); err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(map[string]interface{}{}) // Use "interface{}" instead of "any"
	if err != nil {
		return nil, err
	}

	payload := strings.NewReader(string(marshal))
	if len(params) > 0 {
		payload = strings.NewReader(string(params[0]))
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", res.StatusCode)
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if p.logResponse {
			fn.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *Proxy) Status() bool {
	_, err := p.MakeRequest("/healthy")
	if err != nil {
		return false
	}

	return true
}

func (p *Proxy) WgStatus() ([]byte, error) {
	b, err := p.MakeRequest("/status")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) Start() error {
	_, err := p.MakeRequest("/start")
	if err != nil {
		return err
	}

	return nil
}

func (p *Proxy) Exit() error {
	_, err := p.MakeRequest("/exit")
	if err != nil {
		return err
	}

	return nil
}

func (p *Proxy) Stop() ([]byte, error) {
	b, err := p.MakeRequest("/stop")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) Restart() ([]byte, error) {
	b, err := p.MakeRequest("/restart")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) SetSearchDomain(sd string) ([]byte, error) {
	b, err := p.MakeRequest("/set-search-domain")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) SetDns(sd string) ([]byte, error) {
	b, err := p.MakeRequest("/set-dns")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Proxy) ResetSearchDomain() ([]byte, error) {
	b, err := p.MakeRequest("/restart")
	if err != nil {
		return nil, err
	}

	return b, nil
}
