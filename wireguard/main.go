package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	gatewayPublicKey := os.Getenv("GATEWAY_PUBLIC_KEY")
	hostPublicKey := os.Getenv("HOST_PUBLIC_KEY")
	workspacePublicKey := os.Getenv("WORKSPACE_PUBLIC_KEY")
	privateKey := os.Getenv("PRIVATE_KEY")

	if gatewayPublicKey == "" || hostPublicKey == "" || workspacePublicKey == "" || privateKey == "" {
		panic("missing env vars")
		return
	}

	wgConfig, err := GenerateWireguardConfig(gatewayPublicKey, hostPublicKey, workspacePublicKey, privateKey)
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

	cmd := exec.Command("wg-quick", "up", "wg0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
		return
	}
}

func GenerateWireguardConfig(gatewayPublicKey, hostPublicKey, workspacePublicKey, privateKey string) (string, error) {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 198.18.0.1/32

[Peer]
PublicKey = %s
AllowedIPs = 100.64.0.0/10
Endpoint = gateway:31820
PersistentKeepalive = 25

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.2/32

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.3/32
`, privateKey, gatewayPublicKey, hostPublicKey, workspacePublicKey)

	return config, nil
}
