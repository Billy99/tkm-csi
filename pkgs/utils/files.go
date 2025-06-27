package utils

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-logr/logr"
)

func IsTargetBindMount(target string) (bool, error) {
	dirInfo, err := os.Stat(target)
	if err != nil {
		return false, fmt.Errorf("error getting info for %s: %w", target, err)
	}

	parentDir := filepath.Dir(target)
	parentInfo, err := os.Stat(parentDir)
	if err != nil {
		return false, fmt.Errorf("error getting info for %s: %w", parentDir, err)
	}

	dirSys, ok := dirInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("error getting syscall.Stat_t for %s", target)
	}

	parentSys, ok := parentInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("error getting syscall.Stat_t for %s", parentDir)
	}

	if dirSys.Dev != parentSys.Dev {
		return true, nil
	}

	return false, nil
}

func IsSourceBindMount(namespace, name string, log logr.Logger) (bool, error) {
	cmd := exec.Command("findmnt")
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error executing findmnt: %w", err)
	}

	// Capture and process the output
	output := cmdOutput.String()
	lines := strings.Split(output, "\n")

	// Search for a specific string (e.g., "namespace/name")
	sourcePath := namespace
	sourcePath = filepath.Join(sourcePath, name)
	log.V(1).Info("IsSourceBindMount(): Searching for sourcePath in findmnt output:", "sourcePath", sourcePath)

	found := false
	for _, line := range lines {
		if strings.Contains(line, sourcePath) {
			found = true
			break
		}
	}

	if found {
		log.V(1).Info("IsSourceBindMount(): sourcePath Found", "sourcePath", sourcePath)
	} else {
		log.V(1).Info("IsSourceBindMount(): sourcePath not found", "sourcePath", sourcePath)
	}

	return found, nil
}

// DirSize calculates the total size of a directory and its subdirectories.
func DirSize(path string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(path, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize, err
}
