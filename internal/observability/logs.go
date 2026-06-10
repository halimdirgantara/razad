// Package observability provides log collection and streaming from managed
// application processes to the WebSocket hub.
package observability

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/razad/razad/internal/websocket"
)

// LogStreamer tails log files and pushes new lines to the WebSocket hub.
type LogStreamer struct {
	hub      *websocket.Hub
	dataDir  string
	mu       sync.Mutex
	watching map[string]chan struct{} // app name -> stop channel
}

// NewLogStreamer creates a new log streamer.
func NewLogStreamer(hub *websocket.Hub, dataDir string) *LogStreamer {
	return &LogStreamer{
		hub:      hub,
		dataDir:  dataDir,
		watching: make(map[string]chan struct{}),
	}
}

// WatchApp starts tailing the log file for a specific app.
func (ls *LogStreamer) WatchApp(appName string) {
	ls.mu.Lock()
	if _, ok := ls.watching[appName]; ok {
		ls.mu.Unlock()
		return // already watching
	}
	stop := make(chan struct{})
	ls.watching[appName] = stop
	ls.mu.Unlock()

	go ls.tail(appName, stop)
}

// UnwatchApp stops tailing a specific app's logs.
func (ls *LogStreamer) UnwatchApp(appName string) {
	ls.mu.Lock()
	if stop, ok := ls.watching[appName]; ok {
		close(stop)
		delete(ls.watching, appName)
	}
	ls.mu.Unlock()
}

// tail follows the log file, reading new lines as they're written.
func (ls *LogStreamer) tail(appName string, stop chan struct{}) {
	logPath := filepath.Join(ls.dataDir, "logs", appName, "output.log")

	// Ensure file exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		slog.Warn("logstream: cannot create log dir", "app", appName, "error", err)
		return
	}

	// Open file for reading (create if not exists)
	f, err := os.OpenFile(logPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Warn("logstream: cannot open log file", "app", appName, "error", err)
		return
	}
	defer f.Close()

	// Seek to end to only get new lines
	if _, err := f.Seek(0, 2); err != nil {
		slog.Warn("logstream: seek failed", "app", appName, "error", err)
		return
	}

	reader := bufio.NewReader(f)
	var pending string

	for {
		select {
		case <-stop:
			return
		default:
			chunk, err := reader.ReadString('\n')
			if chunk != "" {
				pending += chunk
			}
			if err != nil {
				// Partial line or no new data yet — wait
				time.Sleep(200 * time.Millisecond)
				continue
			}

			line := strings.TrimRight(pending, "\n\r")
			pending = ""
			if line == "" {
				continue
			}

			ls.hub.BroadcastToApp(appName, websocket.Message{
				Type: "log",
				Payload: map[string]interface{}{
					"app":       appName,
					"message":   line,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			})
		}
	}
}

