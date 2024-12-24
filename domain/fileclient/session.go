package fileclient

import (
	"fmt"
	"os"
	"path"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Session interface {
	GetDevice() (*DeviceData, error)
	SetDevice(dev DeviceData) error

	GetSession() (string, error)
	SetSession(sess string) error

	GetTeam() (string, error)
	GetWsTeam() (string, error)
	SetTeam(team string) error

	GetEnv() (string, error)
	SetEnv(env string) error

	Clear() error

	GetSearchDomain() (string, error)

	Reload() error
}

func (c *fclient) GetDataContext() Session {
	return c.session
}

type DeviceData struct {
	WGconf     string `json:"wg"`
	IpAddress  string `json:"ip"`
	DeviceName string `json:"device"`
}

type EnvCacheData struct{}

type EnvData struct {
	EnvCache EnvCacheData `json:"envCache"`
}

type SessionData struct {
	Session   string                 `json:"session"`
	Team      string                 `json:"team,omitempty"`
	Env       string                 `json:"env,omitempty"`
	TeamsData map[string]*DeviceData `json:"teamsData,omitempty"`
}

type sed struct {
	*SessionData
	handler confighandler.Config[SessionData]
}

func (c *sed) Reload() error {
	sd, err := c.handler.Read()
	if err != nil {
		return err
	}

	c.SessionData = sd
	return nil
}

func (c *sed) GetSearchDomain() (string, error) {
	ed, err := getExtraData()
	if err != nil {
		return "", fn.NewE(err)
	}

	hostsuffix, err := ed.GetDnsHostSuffix()
	if err != nil {
		return "", err
	}

	team, err := c.GetTeam()
	if err != nil {
		return "", nil
	}

	env, err := c.GetEnv()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s.%s", env, team, hostsuffix), nil
}

func (c *sed) Clear() error {
	c.SessionData = &SessionData{}
	return c.handler.Write()
}

func (s *sed) SetDevice(dev DeviceData) error {
	team, err := s.GetTeam()
	if err != nil {
		return err
	}

	if s.TeamsData == nil {
		s.TeamsData = make(map[string]*DeviceData)
	}

	if s.TeamsData[team] == nil {
		s.TeamsData[team] = &DeviceData{}
	}

	s.TeamsData[team] = &dev
	return s.Save()
}

func (s *sed) GetDevice() (*DeviceData, error) {
	team, err := s.GetTeam()
	if err != nil {
		return nil, err
	}

	if s.TeamsData[team] == nil {
		return nil, ErrDeviceNotFound
	}

	return s.TeamsData[team], nil
}

func (s *sed) GetWsTeam() (string, error) {
	if s.Team == "" {
		return "", ErrTeamNotFound
	}

	kt, err := getKlFile()
	if err != nil {
		return "", err
	}

	if kt.TeamName != s.Team {
		if kt.TeamName == "" {
			kt.TeamName = s.Team
			if err := kt.Save(); err != nil {
				return "", err
			}
			return s.Team, nil
		}
		return "", ErrTeamMismatch
	}

	return s.Team, nil
}

func (s *sed) GetTeam() (string, error) {
	if s.Team == "" {
		return "", ErrTeamNotFound
	}

	return s.Team, nil
}

func (s *sed) GetSession() (string, error) {
	if s.Session == "" {
		return "", ErrSessionNotFound
	}

	return s.Session, nil
}

func (s *sed) GetEnv() (string, error) {
	if s.Env == "" {
		return "", ErrEnvNotFound
	}
	return s.Env, nil
}

func (s *sed) SetEnv(env string) error {
	s.Env = env
	return s.Save()
}

func (s *sed) SetTeam(team string) error {
	s.Team = team
	s.Env = ""

	if team != "" {
		configFolder, err := GetConfigFolder()
		if err != nil {
			return err
		}

		tempdir := path.Join(configFolder, team)

		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return nil
		}
	}

	return s.Save()
}

func (s *sed) SetSession(sess string) error {
	s.Session = sess
	return s.Save()
}

func (s *sed) Save() error {
	return s.handler.Write()
}
