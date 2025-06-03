package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func (s *ImageServer) initializeFilesystem() error {
	err := os.MkdirAll(s.cacheDir, 0755)
	if err != nil {
		s.log.Error(err, "error creating directory", "directory", s.cacheDir)
		return err
	}
	s.log.V(1).Info("Successfully created directory", "directory", s.cacheDir)
	return nil
}

func (s *ImageServer) ExtractImage(ctx context.Context, cacheImage, namespace, kernelName string) error {
	// Build command to TCV to Extract OCI Image from URL.
	outputDir := s.cacheDir
	if namespace != "" {
		outputDir = filepath.Join(outputDir, namespace)
	}
	if kernelName != "" {
		outputDir = filepath.Join(outputDir, kernelName)
	}

	loadArgs := []string{"-e", "-i", cacheImage, "-d", outputDir}

	if s.noGpu {
		loadArgs = append(loadArgs, "--no-gpu")
	}

	s.log.V(1).Info("extractImage", "cacheImage", cacheImage, "outputDir", outputDir)

	cmd := exec.CommandContext(ctx, s.tcvBinary, loadArgs...)
	stderr := &bytes.Buffer{}
	cmd.Stdout = io.Discard
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return nil
}
