package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/billy99/tkm-csi/pkgs/constants"
)

func (s *ImageServer) initializeFilesystem() error {
	err := os.MkdirAll(constants.DefaultCacheDir, 0755)
	if err != nil {
		s.log.Error(err, "error creating directory", "directory", constants.DefaultCacheDir)
		return err
	}
	s.log.V(1).Info("Successfully created directory", "directory", constants.DefaultCacheDir)
	return nil
}

func (s *ImageServer) ExtractImage(ctx context.Context, cacheImage, namespace, kernelName string) error {
	// Build command to TCV to Extract OCI Image from URL.
	outputDir := constants.DefaultCacheDir
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

	cmd := exec.CommandContext(ctx, constants.TcvBinary, loadArgs...)
	stderr := &bytes.Buffer{}
	cmd.Stdout = io.Discard
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return nil
}
