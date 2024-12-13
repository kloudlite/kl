package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func klFetch(method string, variables map[string]any, cookie *string, verbose ...bool) ([]byte, error) {
	defer spinner.Client.UpdateMessage("loading please wait")()

	url := constants.ServerURL

	marshal, err := json.Marshal(map[string]any{
		"method": method,
		"args":   []any{variables},
	})
	if err != nil {
		return nil, fn.NewE(err, fmt.Sprintf("failed to marshal apiclient request to server with request %#v on method %s", variables, method))
	}

	payload := strings.NewReader(string(marshal))

	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, fn.NewE(err, fmt.Sprintf("failed to create request while making apiclient request on method %s", method))
	}

	req.Header.Add("authority", "klcli.kloudlite.io")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/json")
	if cookie != nil {
		req.Header.Add("cookie", *cookie)
	}

	// f := spinner.Client.UpdateMessage("loading please wait")
	res, err := http.DefaultClient.Do(req)
	// f()
	if err != nil || res.StatusCode != 200 {
		if err != nil {
			return nil, fn.NewE(err, fmt.Sprintf("failed while making apiclient request to server with method %s", method))
		}

		body, e := io.ReadAll(res.Body)
		if e != nil {
			return nil, e
		}
		return nil, fn.NewE(err, fmt.Sprintf("failed to make apiclient request to server with method %s, status code %d, body %s", method, res.StatusCode, string(body)))
	}

	var respData struct {
		Data   map[string]any
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&respData); err != nil {
		return nil, fn.NewE(err, fmt.Sprintf("failed to read/de-serialize response body for method %s", method))
	}

	if len(respData.Errors) > 0 {
		var errorMessages []string
		for _, e := range respData.Errors {
			if strings.Contains(e.Message, "no rolebinding found") {
				return nil, fn.NewE(fn.Errorf("unauthorized"), fmt.Sprintf("error response from apiclient with method %s", method))
			}
			errorMessages = append(errorMessages, e.Message)
		}

		return nil, fn.NewE(fn.Errorf(strings.Join(errorMessages, "\n")), fmt.Sprintf("error response from apiclient with method %s", method))
	}

	b, err := json.Marshal(respData.Data)
	if err != nil {
		return nil, err
	}

	return b, nil
}
