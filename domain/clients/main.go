package clients

import (
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
)

var Api = func() apiclient.ApiClient {
	ac, err := apiclient.New()
	if err != nil {
		panic(err)
	}

	return ac
}()

var File = func() fileclient.FileClient {
	return Api.GetFClient()
}()
