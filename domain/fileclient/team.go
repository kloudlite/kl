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
		fn.Debug("failed to get config folder")
		return nil, ErrNotFound
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

func (f *fclient) GetDirTeam() (string, error) {
	sd, err := getCtxData()
	if err != nil {
		return "", fn.NewE(err)
	}

	dirTeam, err := sd.GetWsTeam()
	if err != nil {
		if err == ErrTeamNotFound {
			fn.Warn("failed to directory team, trying to get context team")
			return sd.GetTeam()
		}
		return "", fn.NewE(err)
	}

	ctxTeam, err := sd.GetTeam()
	if err != nil {
		if err == ErrNotFound {
			fn.Warn("failed to get context team")
		}
	}

	if ctxTeam != dirTeam {
		fn.Warnf("context team %s is different from directory team %s", ctxTeam, dirTeam)
		fn.Warn("using directory team")
	}

	return dirTeam, nil
}
