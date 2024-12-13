package fileclient

import (
	"os"
	"path"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func getCtxData() (*sed, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, fn.NewE(err, "failed to get config folder")
	}

	chandler := confighandler.GetHandler[SessionData](path.Join(dir, SessionFileName))

	_ = os.MkdirAll(dir, os.ModePerm)

	sd, err := chandler.Read()

	return &sed{
		handler:     chandler,
		SessionData: sd,
	}, nil
}

func getActiveTeamConfigPath() (string, error) {
	sd, err := getCtxData()
	if err != nil {
		return "", err
	}

	s, err := sd.GetTeam()
	if err != nil {
		return "", err
	}

	configFolder, err := GetConfigFolder()
	if err != nil {
		return "", err
	}

	return path.Join(configFolder, s, "config.yml"), nil
}

func (f *fclient) GetTeam() (string, error) {
	sd, err := getCtxData()
	if err != nil {
		return "", fn.NewE(err)
	}

	return sd.GetTeam()
}

func (f *fclient) GetWsTeam() (string, error) {
	sd, err := getCtxData()
	if err != nil {
		return "", fn.NewE(err)
	}

	return sd.GetWsTeam()
}
