package constants

import (
	"fmt"

	"github.com/kloudlite/kl/flags"
)

const (
	DefaultBaseURL              = "https://auth.kloudlite.io"
	RuntimeLinux                = "linux"
	RuntimeDarwin               = "darwin"
	RuntimeWindows              = "windows"
	BashShell                   = "bash"
	FishShell                   = "fish"
	ZshShell                    = "zsh"
	PowerShell                  = "powershell"
	NetworkService              = "Wi-Fi"
	LocalSearchDomains          = ".local"
	NoExistingSearchDomainError = "There aren't any Search Domains set on Wi-Fi."

	ContainerVpnPort = 1729

	DnsServerPort = 5353
)

func baseUrl() string {
	if flags.BaseUrl != "" {
		return flags.BaseUrl
	}
	return DefaultBaseURL
}

var (
	BaseURL = baseUrl()

	LoginUrl = func() string {
		return fmt.Sprintf("%s/cli-login", BaseURL)
	}()
	ServerURL = func() string {
		return fmt.Sprintf("%s/api/", BaseURL)
	}()

	UpdateURL = func() string {
		return "https://kl.kloudlite.io/kloudlite"
	}()
)

var (
	InfraCliName = "kli"
	CoreCliName  = "kl"
)
