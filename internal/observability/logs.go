// Package observability provides log collection and streaming from managed
// application processes to the WebSocket hub.
package observability

import (
	"bufio"
	"fmt"
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

// ReadRecent returns the last `lines` lines from the named log file under
// dataDir/logs/. Returns an empty slice (not an error) when the file does
// not exist yet — a freshly-installed app has no output to show. Lines are
// returned oldest-first within the requested window.
func (ls *LogStreamer) ReadRecent(name string, lines int) ([]string, error) {
	if lines <= 0 {
		lines = 100
	}
	if lines > 5000 {
		lines = 5000
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("observability: log name is required")
	}
	if filepath.IsAbs(name) || strings.Contains(name, "..") {
		return nil, fmt.Errorf("observability: invalid log name %q", name)
	}
	path := filepath.Join(ls.dataDir, "logs", name)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("observability: open log %s: %w", name, err)
	}
	defer f.Close()

	// Read all lines into memory then take the tail. Log files in this
	// service are expected to be modest in size (per-app output); for very
	// large files a ring buffer would be a better fit, but the simple
	// approach keeps the dependency surface small.
	all, err := readAllLines(f)
	if err != nil {
		return nil, fmt.Errorf("observability: read log %s: %w", name, err)
	}
	if len(all) <= lines {
		return all, nil
	}
	return all[len(all)-lines:], nil
}

func readAllLines(f *os.File) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(f)
	// Allow long lines (some app logs include stack traces or JSON payloads).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
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

