package utils

import (
	"log"
	"os"
	"path/filepath"
)

var workspaceDir string

func WorkspaceDir() string {
	if workspaceDir == "" {
		workspaceDir = getWorkspaceDir()
	}

	return workspaceDir
}

func getWorkspaceDir() string {
	var execPath string
	var err error
	if execPath, err = os.Executable(); err != nil {
		log.Fatalf("get exec path fail, %s", err)
	}
	if execPath, err = filepath.EvalSymlinks(execPath); err != nil {
		log.Fatalf("path %s evaluation failed, %s", execPath, err)
	}

	return filepath.Dir(execPath)
}
