package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// CapturarFotoWebcam captura um frame diretamente da câmera via sistema operacional,
// sem passar pelas permissões do navegador ou da WebView.
// Retorna a imagem no formato DataURL (data:image/jpeg;base64,...).
func (a *App) CapturarFotoWebcam() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	if runtime.GOOS == "linux" {
		return capturarFotoLinux(ctx)
	} else if runtime.GOOS == "windows" {
		return capturarFotoWindows(ctx)
	}
	return "", fmt.Errorf("sistema operacional %s não suportado para captura direta", runtime.GOOS)
}

func capturarFotoLinux(ctx context.Context) (string, error) {
	matches, _ := filepath.Glob("/dev/video*")
	if len(matches) == 0 {
		return "", fmt.Errorf("nenhuma câmera encontrada no sistema (/dev/video*)")
	}

	// Usa /dev/video0 por padrão, ou o primeiro encontrado
	dev := "/dev/video0"
	found := false
	for _, m := range matches {
		if m == "/dev/video0" {
			found = true
			break
		}
	}
	if !found && len(matches) > 0 {
		dev = matches[0]
	}

	// 1. Tentar ffmpeg (alta compatibilidade com V4L2)
	if ffmpegPath, err := exec.LookPath("ffmpeg"); err == nil {
		cmd := exec.CommandContext(ctx, ffmpegPath,
			"-y",
			"-f", "v4l2",
			"-video_size", "640x480",
			"-i", dev,
			"-vframes", "1",
			"-f", "image2",
			"-vcodec", "mjpeg",
			"-update", "1",
			"pipe:1",
		)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			b64 := base64.StdEncoding.EncodeToString(stdout.Bytes())
			return "data:image/jpeg;base64," + b64, nil
		}
	}

	// 2. Tentar fswebcam como fallback
	if fswebcamPath, err := exec.LookPath("fswebcam"); err == nil {
		cmd := exec.CommandContext(ctx, fswebcamPath,
			"-d", dev,
			"-r", "640x480",
			"--jpeg", "85",
			"-D", "1",
			"-S", "2",
			"--no-banner",
			"-",
		)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			b64 := base64.StdEncoding.EncodeToString(stdout.Bytes())
			return "data:image/jpeg;base64," + b64, nil
		}
	}

	return "", fmt.Errorf("não foi possível capturar da webcam via Go. Certifique-se de que o ffmpeg está instalado ou use o botão 'Escolher Arquivo'")
}

func capturarFotoWindows(ctx context.Context) (string, error) {
	// 1. Tentar ffmpeg se disponível no PATH
	if ffmpegPath, err := exec.LookPath("ffmpeg"); err == nil {
		cmd := exec.CommandContext(ctx, ffmpegPath,
			"-y",
			"-f", "dshow",
			"-i", "video=Integrated Camera",
			"-vframes", "1",
			"-f", "image2",
			"-vcodec", "mjpeg",
			"-update", "1",
			"pipe:1",
		)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			b64 := base64.StdEncoding.EncodeToString(stdout.Bytes())
			return "data:image/jpeg;base64," + b64, nil
		}
	}

	return "", fmt.Errorf("no Windows, utilize a Câmera do navegador (com permissão automática) ou o botão 'Escolher Arquivo'")
}
