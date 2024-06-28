package server

import (
	"encoding/json"

	"github.com/kloudlite/kl/pkg/functions"
)

func GetCurrentUser() (*User, error) {
	cookie, err := getCookieString()
	if err != nil && cookie == "" {
		return nil, err
	}

	respData, err := klFetch("cli_getCurrentUser", map[string]any{}, &cookie)
	if err != nil {
		return nil, err
	}

	type Resp struct {
		User   User    `json:"data"`
		Errors []error `json:"errors"`
	}

	var resp Resp
	if err = json.Unmarshal(respData, &resp); err != nil {
		return nil, functions.Error(err)
	}

	if len(resp.Errors) > 0 {
		return nil, resp.Errors[0]
	}
	return &resp.User, nil
}
