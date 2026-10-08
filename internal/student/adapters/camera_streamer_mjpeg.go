package adapters

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func (s *CameraStreamer) readFrames(r io.Reader) {
	reader := bufio.NewReader(r)
	for {
		b, err := reader.ReadBytes(0xD8)
		if err != nil {
			break
		}
		if len(b) >= 2 && b[len(b)-2] == 0xFF {
			var frame bytes.Buffer
			frame.WriteByte(0xFF)
			frame.WriteByte(0xD8)
			for {
				chunk, err := reader.ReadBytes(0xD9)
				if err != nil {
					break
				}
				frame.Write(chunk)
				if len(chunk) >= 2 && chunk[len(chunk)-2] == 0xFF {
					s.broadcastFrame(frame.Bytes())
					break
				}
			}
		}
	}
	s.Stop()
}

func (s *CameraStreamer) broadcastFrame(frame []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastFrame = frame
	for ch := range s.clients {
		select {
		case ch <- frame:
		default:
		}
	}
}

func (s *CameraStreamer) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan []byte, 2)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	if !s.running {
		s.startCaptureLocked()
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		if len(s.clients) == 0 {
			s.stopCaptureLocked()
		}
		s.mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	for frame := range ch {
		header := fmt.Sprintf("--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))
		if _, err := w.Write([]byte(header)); err != nil {
			return
		}
		if _, err := w.Write(frame); err != nil {
			return
		}
		if _, err := w.Write([]byte("\r\n")); err != nil {
			return
		}
		if ok {
			flusher.Flush()
		}
	}
}

func (s *CameraStreamer) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	frame := s.lastFrame
	s.mu.RUnlock()
	if len(frame) == 0 {
		http.Error(w, "Sem frame disponível", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(frame)
}
