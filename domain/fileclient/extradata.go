package fileclient

import (
	"fmt"
	"path"
	"time"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
)

type Extra interface {
	Save() error
	GetBackupDns() []string
	SetBackupDns(dns []string) error

	GetBaseUrl() string
	SetBaseUrl(string) error

	SetDnsHostSuffix(suffix string) error
	GetDnsHostSuffix() (string, error)

	GetLastUpdatedCheck() time.Time
	SetLastUpdatedCheck(t time.Time) error
}

type ExtraData struct {
	BaseUrl         string    `json:"baseUrl" yaml:"baseUrl"`
	DnsHostSuffix   string    `json:"dnsHostSuffix" yaml:"dnsHostSuffix"`
	LastUpdateCheck time.Time `json:"lastUpdateCheck" yaml:"lastUpdateCheck"`
	BackUpDns       []string  `json:"backupDns" yaml:"backupDns"`
}

type extra struct {
	*ExtraData
	handler confighandler.Config[ExtraData]
}

func (ed *extra) GetLastUpdatedCheck() time.Time {
	return ed.LastUpdateCheck
}

func (ed *extra) SetLastUpdatedCheck(t time.Time) error {
	ed.LastUpdateCheck = t
	return ed.Save()
}

func (ed *extra) SetDnsHostSuffix(suffix string) error {
	ed.DnsHostSuffix = suffix
	return ed.Save()
}

func (ed *extra) GetDnsHostSuffix() (string, error) {
	if ed.DnsHostSuffix == "" {
		return "", fmt.Errorf("dns host suffix is empty")
	}

	return ed.DnsHostSuffix, nil
}

func (ed *extra) GetBaseUrl() string {
	return ed.BaseUrl
}

func (ed *extra) SetBaseUrl(url string) error {
	ed.BaseUrl = url
	return ed.Save()
}

func (ed *extra) SetBackupDns(dns []string) error {
	ed.BackUpDns = dns
	return ed.Save()
}

func (ed *extra) GetBackupDns() []string {
	return ed.BackUpDns
}

func (ed *extra) Save() error {
	return ed.handler.Write()
}

func (fc *fclient) GetExtraData() (Extra, error) {
	return getExtraData()
}

func getExtraData() (Extra, error) {
	cdir, err := GetConfigFolder()
	if err != nil {
		return nil, err
	}

	chandler := confighandler.GetHandler[ExtraData](path.Join(cdir, ExtraDataFileName))

	wcd, _ := chandler.Read()

	resp := extra{
		ExtraData: wcd,
		handler:   chandler,
	}

	return &resp, nil
}
