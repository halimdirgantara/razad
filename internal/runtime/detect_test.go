package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDir creates a temporary directory with the given files.
func setupTestDir(t *testing.T, files []string) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", f, err)
		}
	}
	return dir
}

func TestDetect_NodeJS(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"package.json", "index.js"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "node" {
		t.Errorf("expected node, got %s", result.Name)
	}
	if result.Confidence != "high" {
		t.Errorf("expected high confidence, got %s", result.Confidence)
	}
	if result.StartCommand != "npm start" {
		t.Errorf("expected 'npm start', got %s", result.StartCommand)
	}
}

func TestDetect_Go(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"go.mod", "main.go"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "go" {
		t.Errorf("expected go, got %s", result.Name)
	}
	if result.Confidence != "high" {
		t.Errorf("expected high confidence, got %s", result.Confidence)
	}
}

func TestDetect_PythonRequirements(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"requirements.txt"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "python" {
		t.Errorf("expected python, got %s", result.Name)
	}
}

func TestDetect_PythonPyProject(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"pyproject.toml"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "python" {
		t.Errorf("expected python, got %s", result.Name)
	}
}

func TestDetect_UnknownProject(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"README.md", "LICENSE"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "unknown" {
		t.Errorf("expected unknown, got %s", result.Name)
	}
	if result.Confidence != "low" {
		t.Errorf("expected low confidence, got %s", result.Confidence)
	}
}

func TestDetect_EmptyDirectory(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{})

	// An empty directory returns "unknown" with low confidence
	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect should not error on empty dir: %v", err)
	}
	if result.Name != "unknown" {
		t.Errorf("expected unknown for empty dir, got %s", result.Name)
	}
}

func TestDetect_AmbiguousMultipleSignals(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"package.json", "go.mod"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Confidence != "ambiguous" {
		t.Errorf("expected ambiguous confidence, got %s", result.Confidence)
	}
}

func TestDetect_WithOverride(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"some-random-file"})

	// Override forces a specific runtime regardless of project contents
	result, err := d.Detect(dir, "python")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "python" {
		t.Errorf("expected python override, got %s", result.Name)
	}
	if result.Confidence != "high" {
		t.Errorf("expected high confidence for override, got %s", result.Confidence)
	}
}

func TestDetect_Dockerfile(t *testing.T) {
	d := New()
	dir := setupTestDir(t, []string{"Dockerfile"})

	result, err := d.Detect(dir, "")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result.Name != "container" {
		t.Errorf("expected container, got %s", result.Name)
	}
	if result.Confidence != "low" {
		t.Errorf("expected low confidence, got %s", result.Confidence)
	}
}

func TestLookupStartCommand(t *testing.T) {
	tests := []struct {
		runtime string
		want    string
	}{
		{"node", "npm start"},
		{"go", "go run ."},
		{"python", "python main.py"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		got := LookupStartCommand(tt.runtime)
		if got != tt.want {
			t.Errorf("LookupStartCommand(%q) = %q, want %q", tt.runtime, got, tt.want)
		}
	}
}
