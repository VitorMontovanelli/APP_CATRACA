package adapters

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"catraca-app/internal/student/ports"
)

var v4l2DeviceRegex = regexp.MustCompile(`^/dev/video[0-9]+$`)

type HardwareCameraAdapter struct{}

func NewHardwareCameraAdapter() ports.CameraPort {
	return &HardwareCameraAdapter{}
}

func (a *HardwareCameraAdapter) CaptureFrame(ctx context.Context) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	if runtime.GOOS == "linux" {
		return a.captureLinux(timeoutCtx)
	} else if runtime.GOOS == "windows" {
		return a.captureWindows(timeoutCtx)
	}
	return "", fmt.Errorf("sistema operacional %s não suportado para captura direta de câmera", runtime.GOOS)
}

func (a *HardwareCameraAdapter) captureLinux(ctx context.Context) (string, error) {
	matches, _ := filepath.Glob("/dev/video*")
	if len(matches) == 0 {
		return "", fmt.Errorf("nenhuma câmera encontrada no sistema (/dev/video*)")
	}

	dev := "/dev/video0"
	for _, m := range matches {
		if v4l2DeviceRegex.MatchString(m) {
			dev = m
			break
		}
	}

	// 1. Tentar ffmpeg com V4L2
	if ffmpegPath, err := exec.LookPath("ffmpeg"); err == nil {
		cmd := exec.CommandContext(ctx, ffmpegPath,
			"-y", "-f", "v4l2", "-video_size", "640x480",
			"-i", dev, "-vframes", "1", "-f", "image2",
			"-vcodec", "mjpeg", "-update", "1", "pipe:1",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(stdout.Bytes()), nil
		}
	}

	// 2. Tentar fswebcam como fallback
	if fswebcamPath, err := exec.LookPath("fswebcam"); err == nil {
		cmd := exec.CommandContext(ctx, fswebcamPath,
			"-d", dev, "-r", "640x480", "--jpeg", "85",
			"-D", "1", "-S", "2", "--no-banner", "-",
		)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(stdout.Bytes()), nil
		}
	}

	return "", fmt.Errorf("não foi possível capturar da webcam via hardware. Instale o ffmpeg no sistema")
}

func (a *HardwareCameraAdapter) captureWindows(ctx context.Context) (string, error) {
	if ffmpegPath, err := exec.LookPath("ffmpeg"); err == nil {
		cmd := exec.CommandContext(ctx, ffmpegPath,
			"-y", "-f", "dshow", "-i", "video=Integrated Camera",
			"-vframes", "1", "-f", "image2", "-vcodec", "mjpeg",
			"-update", "1", "pipe:1",
		)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(stdout.Bytes()), nil
		}
	}
	return "", fmt.Errorf("captura via ffmpeg indisponível no Windows")
}
