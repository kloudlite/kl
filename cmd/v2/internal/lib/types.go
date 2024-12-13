package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"
)

type KLConfig struct {
	ConfigFile string `json:"-"`

	Version    string `json:"version" yaml:"version"`
	DefaultEnv string `json:"defaultEnv,omitempty" yaml:"defaultEnv,omitempty"`
	TeamName   string `json:"teamName,omitempty" yaml:"teamName,omitempty"`

	Packages  []string `json:"packages" yaml:"packages"`
	Libraries []string `json:"libraries" yaml:"libraries"`

	EnvVars EnvVars `json:"envVars" yaml:"envVars"`
	Mounts  Mounts  `json:"mounts" yaml:"mounts"`
	Ports   []int   `json:"ports" yaml:"ports"`

	packagesMap  map[string]int `json:"-"`
	librariesMap map[string]int `json:"-"`
}

var klCfgTemplate *template.Template

func createSet[T comparable](v []T) []T {
	m := make(map[T]struct{}, len(v))
	result := make([]T, 0, len(v))

	for i := range v {
		if _, ok := m[v[i]]; !ok {
			m[v[i]] = struct{}{}
			result = append(result, v[i])
		}
	}

	return result
}

func skipDefault[T comparable](v []T) []T {
	result := make([]T, 0, len(v))
	var dv T
	for i := range v {
		if v[i] == dv {
			// INFO: v[i] is the default value
			continue
		}
		result = append(result, v[i])
	}
	return result
}

func init() {
	klConfigTemplate := /*gotmpl*/ `
{{- with . -}}
version: {{ .Version | quote }}
defaultEnv: {{ .DefaultEnv | quote }}

# you, can add packages from search.nixos.org
{{- $fpackages := .Packages | skipDefaults }}
packages: {{- if $fpackages }} {{ $fpackages | toYAML | nindent 2}} {{- else }} [] {{- "\n" -}} {{- end }}

{{- $flibraries := .Libraries | skipDefaults }}
libraries: {{- if $flibraries }} {{ $flibraries | toYAML | nindent 2}} {{- else }} [] {{- "\n" -}} {{- end }}

{{- $fenvVars := .EnvVars | skipDefaults }}
envVars: {{- if $fenvVars }} {{ $fenvVars | toYAML | nindent 2}} {{- else }} [] {{- "\n" -}} {{- end }}

{{- $fmounts := .Mounts | skipDefaults }}
mounts: {{- if $fmounts }} {{ $fmounts | toYAML | nindent 2}} {{- else }} [] {{- "\n" -}} {{- end }}

ports: []
teamName: {{.TeamName | quote }}
{{- end }}
`

	klCfgTemplate = template.New("kl-config")

	indent := func(spaces int, v string) string {
		pad := strings.Repeat(" ", spaces)
		return pad + strings.Replace(v, "\n", "\n"+pad, -1)
	}

	m := map[string]any{
		"toYAML": func(v any) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			b2, err := yaml.JSONToYAML(b)
			if err != nil {
				return "", err
			}

			return string(b2), nil
		},
		"skipDefaults": func(v any) any {
			switch pv := v.(type) {
			case []string:
				{
					return skipDefault(pv)
				}
			case []int:
				{
					return skipDefault(pv)
				}
			default:
				return v
			}
		},

		"quote": func(v string) string {
			return fmt.Sprintf("%q", v)
		},

		"uniq": func(v any) any {
			switch pv := v.(type) {
			case []string:
				{
					return createSet(pv)
				}
			case []int:
				{
					return createSet(pv)
				}
			default:
				return v
			}
		},
		"indent": indent,
		"nindent": func(spaces int, v string) string {
			return "\n" + indent(spaces, v)
		},
	}

	klCfgTemplate.Funcs(m)
	var err error
	klCfgTemplate, err = klCfgTemplate.Parse(klConfigTemplate)
	if err != nil {
		panic(err)
	}
}

func (klc *KLConfig) AddEnvVar(v EnvType) error {
	// FIXME: should not be able to add a single key multiple times
	klc.EnvVars = append(klc.EnvVars, v)

	b := new(bytes.Buffer)
	if err := klCfgTemplate.ExecuteTemplate(b, "kl-config", klc); err != nil {
		return err
	}

	return os.WriteFile(klc.ConfigFile, b.Bytes(), 0o700)
}

type parsedPackage struct {
	Name        string
	Version     *string
	NixpkgsHash string
}

func (p *parsedPackage) String() string {
	if p.Version != nil {
		return fmt.Sprintf("%s@%s|%s", p.Name, *p.Version, p.NixpkgsHash)
	}
	return fmt.Sprintf("%s|%s", p.Name, p.NixpkgsHash)
}

func ParsePackage(pkg string) (*parsedPackage, error) {
	return parsePackage(pkg)
}

func parsePackage(pkg string) (*parsedPackage, error) {
	sp := strings.SplitN(pkg, "|", 2)

	pp := parsedPackage{}

	switch len(sp) {
	case 1:
		pp.NixpkgsHash = sp[0]
		nixtag := strings.SplitN(sp[0], "#", 2)
		if len(nixtag) != 2 {
			return nil, fmt.Errorf("invalid package %s", pkg)
		}
		pp.Name = nixtag[1]
		pp.Version = nil
	case 2:
		pp.NixpkgsHash = sp[1]
		pp.Name = sp[0]
		nsp := strings.SplitN(sp[0], "@", 2)
		if len(nsp) == 2 {
			pp.Name = nsp[0]
			pp.Version = &nsp[1]
		}
	}

	return &pp, nil
}

func (klc *KLConfig) AddPackage(v ...string) error {
	for _, pkg := range v {
		pp, err := parsePackage(pkg)
		if err != nil {
			return err
		}

		if idx, ok := klc.packagesMap[pp.Name]; ok {
			klc.Packages[idx] = pp.String()
			continue
		}
		klc.packagesMap[pp.Name] = len(klc.Packages)
		klc.Packages = append(klc.Packages, pp.String())
	}

	return klc.save()
}

func (klc *KLConfig) RemovePackage(v ...string) error {
	for _, pkg := range v {
		pp, err := parsePackage(pkg)
		if err != nil {
			return err
		}

		if idx, ok := klc.packagesMap[pp.Name]; ok {
			klc.Packages[idx] = ""
			continue
		}
	}

	return klc.save()
}

func (klc *KLConfig) AddLibrary(v ...string) error {
	for _, pkg := range v {
		pp, err := parsePackage(pkg)
		if err != nil {
			return err
		}

		if idx, ok := klc.librariesMap[pp.Name]; ok {
			klc.Libraries[idx] = pp.String()
			continue
		}
		klc.librariesMap[pp.Name] = len(klc.Libraries)
		klc.Libraries = append(klc.Libraries, pp.String())
	}

	return klc.save()
}

func (klc *KLConfig) RemoveLibrary(v ...string) error {
	for _, pkg := range v {
		pp, err := parsePackage(pkg)
		if err != nil {
			return err
		}

		if idx, ok := klc.librariesMap[pp.Name]; ok {
			klc.Libraries[idx] = ""
			continue
		}
	}

	return klc.save()
}

func (klc *KLConfig) AddMount(v Mount) error {
	klc.Mounts = append(klc.Mounts, v)
	return klc.save()
}

func (klc *KLConfig) save() error {
	b := new(bytes.Buffer)
	if err := klCfgTemplate.ExecuteTemplate(b, "kl-config", klc); err != nil {
		return err
	}

	return os.WriteFile(klc.ConfigFile, b.Bytes(), 0o700)
}

type EnvType struct {
	Key       string  `json:"key" yaml:"key"`
	Value     *string `json:"value,omitempty" yaml:"value,omitempty"`
	ConfigRef *string `json:"configRef,omitempty" yaml:"configRef,omitempty"`
	SecretRef *string `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
	MresRef   *string `json:"mresRef,omitempty" yaml:"mresRef,omitempty"`
}

type ParsedKLConfig struct {
	ConfigFile string `json:"configFile"`

	CacheFile string `json:"cacheFile"`
	CacheDir  string `json:"cacheDir"`

	Packages  []string `json:"packages"`
	Libraries []string `json:"libraries"`

	EnvVars []string          `json:"envVars"`
	Mounts  map[string][]byte `json:"mounts"`

	Hash string `json:"hash"`
}

type Mount struct {
	Path      string  `json:"path"`
	ConfigRef *string `json:"configRef,omitempty" yaml:"configRef,omitempty"`
	SecretRef *string `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
}

type (
	Mounts  []Mount
	EnvVars []EnvType
)
