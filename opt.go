package tnnlr

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

/*
Bookkeeping operations for tnnlr.
Logging and pid tracking for launched processes.
*/

var baseDir = "~/.tnnlr"

var relProc = "proc"
var relLog = "log"

func getRelativePath(subdir string) (string, error) {
	basePath, err := homedir.Expand(baseDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, subdir), nil
}

func createRelDir(subdir string) error {
	procDir, err := getRelativePath(subdir)
	if err != nil {
		return err
	}
	return os.MkdirAll(procDir, os.ModePerm)
}
