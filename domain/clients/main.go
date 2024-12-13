package clients

import (
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

var Api = apiclient.Api
var File = fileclient.File
var Spinner = spinner.Client
