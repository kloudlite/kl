package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kloudlite/kl/klbox-docker/devboxfile"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	var configFile string
	flag.StringVar(&configFile, "conf", "", "--conf /path/to/config.json")
	flag.Parse()

	if configFile == "" {
		return fmt.Errorf("no config file provided")
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var c devboxfile.DevboxConfig
	err = json.Unmarshal(b, &c)
	if err != nil {
		return err
	}

	for k, v := range c.KlConfig.Mounts {
		if err := os.MkdirAll(filepath.Dir(k), fs.ModePerm); err != nil {
			return err
		}

		if err := os.Chown(filepath.Dir(k), 1000, 1000); err != nil {
			return err
		}

		if err := os.WriteFile(k, []byte(v), fs.ModePerm); err != nil {
			return err
		}

		if err := os.Chown(k, 1000, 1000); err != nil {
			return err
		}
	}

	wgPath := "/etc/resolv.conf"
	if c.KlConfig.Dns != "" {
		if err := os.WriteFile(wgPath, []byte(c.KlConfig.Dns), fs.ModePerm); err != nil {
			return err
		}
	}

	for _, v := range c.KlConfig.InitScripts {
		if err := RunScript(fmt.Sprintf("bash -c %q", v)); err != nil {
			fn.PrintError(fmt.Errorf("error running init script: %q", v))
		}
	}

	return nil
}

func RunScript(script string) error {
	r := csv.NewReader(strings.NewReader(script))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)

	fn.Log("[#] " + strings.Join(cmdArr, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = "/home/kl/workspace"
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return err
}
