package fileclient

import (
	"fmt"
	"math/rand"
	"os"
	"path"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Session interface {
	GetK3sPort() (*string, error)
	SetK3sPort(port string) error

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

type TeamData struct {
	Device  *DeviceData `json:"device"`
	K3sPort *string     `json:"k3sPort"`
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
	Session   string               `json:"session"`
	Team      string               `json:"team,omitempty"`
	Env       string               `json:"env,omitempty"`
	TeamsData map[string]*TeamData `json:"teamsData,omitempty"`
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

func (s *sed) SetK3sPort(port string) error {
	team, err := s.GetTeam()
	if err != nil {
		return err
	}

	if s.TeamsData == nil {
		s.TeamsData = make(map[string]*TeamData)
	}

	if s.TeamsData[team] == nil {
		s.TeamsData[team] = &TeamData{}
	}

	s.TeamsData[team].K3sPort = &port
	return s.Save()
}

func (s *sed) GetK3sPort() (*string, error) {
	team, err := s.GetTeam()
	if err != nil {
		return nil, err
	}

	if s.TeamsData[team] == nil {
		s.TeamsData[team] = &TeamData{}
	}

	if s.TeamsData[team].K3sPort == nil {

		count := 0
		for {
			count++
			if count > 10 {
				return nil, ErrK3sPortNotFound
			}

			i := rand.Intn(100) + 33000
			if fn.IsPortFree(fmt.Sprintf("%d", i)) {
				if err := s.SetK3sPort(fmt.Sprintf("%d", i)); err != nil {
					return nil, err
				}

				return s.GetK3sPort()
			}
		}
	}

	return s.TeamsData[team].K3sPort, nil
}

func (s *sed) SetDevice(dev DeviceData) error {
	team, err := s.GetTeam()
	if err != nil {
		return err
	}

	if s.TeamsData == nil {
		s.TeamsData = make(map[string]*TeamData)
	}

	if s.TeamsData[team] == nil {
		s.TeamsData[team] = &TeamData{}
	}

	s.TeamsData[team].Device = &dev
	return s.Save()
}

func (s *sed) GetDevice() (*DeviceData, error) {
	team, err := s.GetTeam()
	if err != nil {
		return nil, err
	}

	if s.TeamsData[team] == nil {
		s.TeamsData[team] = &TeamData{}
		if err := s.Save(); err != nil {
			return nil, err
		}
	}

	if s.TeamsData[team].Device == nil {
		return nil, ErrDeviceNotFound
	}

	return s.TeamsData[team].Device, nil
}

func (s *sed) GetWsTeam() (string, error) {
	kt, err := getKlFile()
	if err != nil {
		return "", err
	}

	if kt.TeamName == "" {
		return "", ErrTeamNotFound
	}

	return kt.TeamName, nil
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
