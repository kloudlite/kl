package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {

	hostPublicKey := os.Getenv("HOST_PUBLIC_KEY")
	workspacePublicKey := os.Getenv("WORKSPACE_PUBLIC_KEY")
	privateKey := os.Getenv("PRIVATE_KEY")

	if hostPublicKey == "" || workspacePublicKey == "" || privateKey == "" {
		panic("missing env vars")
		return
	}

	wgConfig, err := GenerateWireguardConfig(hostPublicKey, workspacePublicKey, privateKey)
	if err != nil {
		panic(err)
		return
	}

	wfPath := "/etc/wireguard"
	if err := os.MkdirAll(wfPath, os.ModePerm); err != nil {
		panic(err)
		return
	}

	f, err := os.Create("/etc/wireguard/wg0.conf")
	if err != nil {
		panic(err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(wgConfig)
	if err != nil {
		panic(err)
		return
	}

	cmdDown := exec.Command("wg-quick", "down", "wg0")
	err = cmdDown.Run()
	if err != nil {
		// ignore error to down wireguard
	}
	cmd := exec.Command("wg-quick", "up", "wg0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
		return
	}

	err = http.ListenAndServe("0.0.0.0:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := exec.Command("sh", "-c", "/usr/local/bin/healthcheck.sh").Output()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			// delete pod using k8s client

			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	if err != nil {
		panic(err)
		return
	}

	//sigChan := make(chan os.Signal, 1)
	//signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	//<-sigChan
}

func GenerateWireguardConfig(hostPublicKey, workspacePublicKey, privateKey string) (string, error) {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 198.18.0.1/32
ListenPort = 31820
PostUp = iptables -t nat -I POSTROUTING -o kloudlite-wg -j MASQUERADE

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.2/32

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.3/32
`, privateKey, hostPublicKey, workspacePublicKey)
	return config, nil
}
