package adapters

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type CameraStreamer struct {
	mu        sync.RWMutex
	listener  net.Listener
	server    *http.Server
	port      int
	cmd       *exec.Cmd
	running   bool
	lastFrame []byte
	clients   map[chan []byte]struct{}
}

var (
	globalStreamer *CameraStreamer
	streamerOnce   sync.Once
)

func GetCameraStreamer() *CameraStreamer {
	streamerOnce.Do(func() {
		globalStreamer = &CameraStreamer{
			clients: make(map[chan []byte]struct{}),
		}
		_ = globalStreamer.initServer()
	})
	return globalStreamer
}

func (s *CameraStreamer) initServer() error {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.listener = l
	s.port = l.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", s.handleStream)
	mux.HandleFunc("/snapshot", s.handleSnapshot)

	s.server = &http.Server{Handler: mux}
	go s.server.Serve(l)
	return nil
}

func (s *CameraStreamer) GetStreamURL() string {
	s.mu.Lock()
	if !s.running {
		s.startCaptureLocked()
	}
	s.mu.Unlock()
	return fmt.Sprintf("http://127.0.0.1:%d/stream?t=%d", s.port, time.Now().UnixMilli())
}

func (s *CameraStreamer) findDevice() string {
	matches, _ := filepath.Glob("/dev/video*")
	dev := "/dev/video0"
	for _, m := range matches {
		if v4l2DeviceRegex.MatchString(m) {
			dev = m
			break
		}
	}
	return dev
}

func (s *CameraStreamer) startCaptureLocked() {
	if s.running {
		return
	}
	dev := s.findDevice()
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("ffmpeg",
			"-f", "v4l2", "-video_size", "640x480",
			"-i", dev, "-f", "mjpeg", "-q:v", "5", "pipe:1",
		)
	} else {
		cmd = exec.Command("ffmpeg",
			"-f", "dshow", "-i", "video=Integrated Camera",
			"-f", "mjpeg", "-q:v", "5", "pipe:1",
		)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}
	s.cmd = cmd
	s.running = true
	go s.readFrames(stdout)
}

func (s *CameraStreamer) CaptureSnapshot() (string, error) {
	s.mu.RLock()
	frame := s.lastFrame
	s.mu.RUnlock()
	if len(frame) == 0 {
		return "", fmt.Errorf("sem imagem capturada no momento")
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(frame), nil
}

func (s *CameraStreamer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopCaptureLocked()
}

func (s *CameraStreamer) stopCaptureLocked() {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		s.cmd = nil
	}
	s.running = false
}
