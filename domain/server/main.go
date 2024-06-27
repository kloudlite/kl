package server

import (
	"encoding/json"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
)

type User struct {
	UserId string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

var authSecret string

func getCookie(options ...functions.Option) (string, error) {
	return fileclient.GetCookieString(options...)
}

type Response[T any] struct {
	Data   T       `json:"data"`
	Errors []error `json:"errors"`
}

func GetFromResp[T any](respData []byte) (*T, error) {
	var resp Response[T]
	err := json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, functions.NewE(err)
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
		return nil, functions.NewE(err)
	}

	var data []T
	for _, v := range resp.Edges {
		data = append(data, v.Node)
	}

	return data, nil
}
