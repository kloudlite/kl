package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

const (
	StatusFailed = "failed to get status"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "get status of your current context (user, account, environment, vpn status)",
	Run: func(cmd *cobra.Command, _ []string) {

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if u, err := apic.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		acc, err := fc.CurrentAccountName()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Account: ")), acc))
		}

		e, err := apic.EnsureEnv()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
		} else if errors.Is(err, fileclient.NoEnvSelected) {
			filePath := fn.ParseKlFile(cmd)
			klFile, err := fc.GetKlFile(filePath)
			if err != nil {
				fn.PrintError(err)
				return
			}
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), klFile.DefaultEnv))
		}

		err = getK3sStatus()
		if err != nil {
			return
		}

	},
}

func getK3sStatus() error {
	resp, err := http.Get(fmt.Sprintf("http://%s:8080/apis/apps/v1/namespaces/kl-gateway/deployments/default", constants.K3sServerIp))
	if err != nil {
		return fn.NewE(err, StatusFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fn.NewE(err, StatusFailed)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var data struct {
		Status struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		} `json:"status"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	isReady := false

	for _, c := range data.Status.Conditions {
		if c.Type == "Available" && c.Status == "True" {
			isReady = true
			break
		}
	}

	if isReady {
		fn.Log(fmt.Sprint("kloudlite gateway: ", text.Green("ready")))
	} else {
		fn.Log(fmt.Sprint("kloudlite gateway: ", text.Yellow("not ready")))
	}

	resp, err = http.Get(fmt.Sprintf("http://%s:8080/apis/apps/v1/namespaces/kloudlite/deployments/kl-agent", constants.K3sServerIp))
	if err != nil {
		return fn.NewE(err, StatusFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fn.NewE(err, StatusFailed)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	isReady = false

	for _, c := range data.Status.Conditions {
		if c.Type == "Available" && c.Status == "True" {
			isReady = true
			break
		}
	}

	if isReady {
		fn.Log(fmt.Sprint("kloudlite agent: ", text.Green("ready")))
	} else {
		fn.Log(fmt.Sprint("kloudlite agent: ", text.Yellow("not ready")))
	}

	resp, err = http.Get(fmt.Sprintf("http://%s:8080/apis/apps/v1/namespaces/kloudlite/deployments/kl-agent-operator", constants.K3sServerIp))
	if err != nil {
		return fn.NewE(err, StatusFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fn.NewE(err, StatusFailed)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	isReady = false

	for _, c := range data.Status.Conditions {
		if c.Type == "Available" && c.Status == "True" {
			isReady = true
			break
		}
	}

	if isReady {
		fn.Log(fmt.Sprint("kloudlite agent operator: ", text.Green("ready")))
	} else {
		fn.Log(fmt.Sprint("kloudlite agent operator: ", text.Yellow("not ready")))
	}

	return nil
}
