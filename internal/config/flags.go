package config

import (
	"strings"
)

type LeGrimoireFlags struct {
	DataDir string
	Host    string
	Port    int
}

func (f *LeGrimoireFlags) Clean() {
	f.DataDir = strings.TrimSpace(f.DataDir)
	f.Host = strings.TrimSpace(f.Host)
}
