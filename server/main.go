package server

import (
	"encoding/json"
	"net/http"
	"time"

	fn "github.com/kloudlite/kl/pkg/functions"
	nanoid "github.com/matoous/go-nanoid/v2"
)

type User struct {
	UserId string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

var authSecret string

func CreateRemoteLogin() (loginId string, err error) {
	authSecret, err = nanoid.New(32)
	if err != nil {
		return "", err
	}

	respData, err := klFetch("cli_createRemoteLogin", map[string]any{
		"secret": authSecret,
	}, nil)
	if err != nil {
		return "", err
	}

	type Response struct {
		Id string `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

func Login(loginId string) error {
	for {
		respData, err := klFetch("cli_getRemoteLogin", map[string]any{
			"loginId": loginId,
			"secret":  authSecret,
		}, nil)

		if err != nil {
			return fn.Error(err)
		}
		type Response struct {
			RemoteLogin struct {
				Status     string `json:"status"`
				AuthHeader string `json:"authHeader"`
			} `json:"data"`
		}

		var loginStatusResponse Response
		if err = json.Unmarshal(respData, &loginStatusResponse); err != nil {
			return fn.Error(err)
		}

		if loginStatusResponse.RemoteLogin.Status == "succeeded" {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("Cookie", loginStatusResponse.RemoteLogin.AuthHeader)
			cookie, _ := req.Cookie("hotspot-session")

			return SaveAuthSession(cookie.Value)
		}

		if loginStatusResponse.RemoteLogin.Status == "failed" {
			return fn.Error(err, "remote login failed")
		}

		if loginStatusResponse.RemoteLogin.Status == "pending" {
			time.Sleep(time.Second * 2)
			continue
		}
	}
}

type Response[T any] struct {
	Data   T       `json:"data"`
	Errors []error `json:"errors"`
}

func GetFromResp[T any](respData []byte) (*T, error) {
	var resp Response[T]
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, err
	}

	if len(resp.Errors) > 0 {
		return nil, resp.Errors[0]
	}
	return &resp.Data, nil
}

type ItemList[T any] struct {
	Edges Edges[T] `json:"edges"`
}

func GetFromRespForEdge[T any](respData []byte) ([]T, error) {
	resp, err := GetFromResp[ItemList[T]](respData)
	if err != nil {
		return nil, err
	}

	var data []T
	for _, v := range resp.Edges {
		data = append(data, v.Node)
	}

	return data, nil
}
