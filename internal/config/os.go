package config

import (
	"os"
	"path/filepath"
)

var (
	OsReadFile  = os.ReadFile
	OsWriteFile = os.WriteFile
	OsStat      = os.Stat
	OsMkdirAll  = os.MkdirAll
	OsRemove    = os.Remove
	OsReadDir   = os.ReadDir
	OsOpen      = os.Open
	OsOpenFile  = os.OpenFile

	FilepathAbs = filepath.Abs
)
