package lib

import (
	"crypto/md5"
	"fmt"
	"slices"
)

type GenHashArgs struct {
	EnvVars   []string
	Mounts    map[string][]byte
	Packages  []string
	Libraries []string
}

func GenHash(args GenHashArgs) string {
	h := md5.New()

	slices.Sort(args.Packages)
	for i := range args.Packages {
		h.Write([]byte(args.Packages[i]))
	}

	slices.Sort(args.Libraries)
	for i := range args.Libraries {
		h.Write([]byte(args.Libraries[i]))
	}

	slices.Sort(args.EnvVars)
	for i := range args.EnvVars {
		h.Write([]byte(args.EnvVars[i]))
	}

	mk := make([]string, 0, len(args.Mounts))
	for k := range args.Mounts {
		mk = append(mk, k)
	}

	for _, k := range mk {
		h.Write(args.Mounts[k])
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
