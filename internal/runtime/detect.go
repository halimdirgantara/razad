// Package runtime detects application runtimes from project file structure.
package runtime

import (
	"os"
	"strings"
)

// RuntimeResult describes a detected application runtime.
type RuntimeResult struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	StartCommand string `json:"start_command"`
	Confidence   string `json:"confidence"`
	BuildCommand string `json:"build_command,omitempty"`
}

// Detector inspects a project directory and returns the inferred runtime.
type Detector struct{}

// New creates a new runtime detector.
func New() *Detector {
	return &Detector{}
}

// Detect examines the project directory and returns runtime information.
func (d *Detector) Detect(projectDir string, override string) (RuntimeResult, error) {
	if override != "" {
		return resultFromOverride(override), nil
	}

	files, err := readDir(projectDir)
	if err != nil {
		return RuntimeResult{}, err
	}

	var results []RuntimeResult
	hasSignal := false

	for _, f := range files {
		switch f {
		case "package.json":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "node", DisplayName: "Node.js",
				StartCommand: "npm start",
				BuildCommand: "npm install",
				Confidence:   "high",
			})
		case "bun.lockb", "bunfig.toml":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "bun", DisplayName: "Bun",
				StartCommand: "bun run start",
				BuildCommand: "bun install",
				Confidence:   "high",
			})
		case "composer.json":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "php", DisplayName: "PHP",
				StartCommand: "php artisan serve",
				BuildCommand: "composer install",
				Confidence:   "high",
			})
		case "go.mod":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "go", DisplayName: "Go",
				StartCommand: "go run .",
				BuildCommand: "go build -o /dev/null",
				Confidence:   "high",
			})
		case "requirements.txt":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "python", DisplayName: "Python",
				StartCommand: "python main.py",
				BuildCommand: "pip install -r requirements.txt",
				Confidence:   "medium",
			})
		case "pyproject.toml":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "python", DisplayName: "Python",
				StartCommand: "python main.py",
				BuildCommand: "pip install .",
				Confidence:   "medium",
			})
		case "Gemfile":
			hasSignal = true
			results = append(results, RuntimeResult{
				Name: "ruby", DisplayName: "Ruby",
				StartCommand: "bundle exec ruby main.rb",
				BuildCommand: "bundle install",
				Confidence:   "medium",
			})
		case "Dockerfile":
			results = append(results, RuntimeResult{
				Name: "container", DisplayName: "Container (Docker)",
				Confidence: "low",
			})
		}
	}

	if len(results) == 0 {
		return RuntimeResult{Name: "unknown", DisplayName: "Unknown", Confidence: "low"}, nil
	}

	if !hasSignal {
		return results[0], nil
	}

	if len(results) > 1 {
		highCount := 0
		for _, r := range results {
			if r.Confidence == "high" {
				highCount++
			}
		}
		if highCount > 1 {
			results[0].Confidence = "ambiguous"
		}
	}

	return results[0], nil
}

func resultFromOverride(name string) RuntimeResult {
	switch strings.ToLower(name) {
	case "node":
		return RuntimeResult{Name: "node", DisplayName: "Node.js", StartCommand: "npm start", BuildCommand: "npm install", Confidence: "high"}
	case "bun":
		return RuntimeResult{Name: "bun", DisplayName: "Bun", StartCommand: "bun run start", BuildCommand: "bun install", Confidence: "high"}
	case "php":
		return RuntimeResult{Name: "php", DisplayName: "PHP", StartCommand: "php artisan serve", BuildCommand: "composer install", Confidence: "high"}
	case "go":
		return RuntimeResult{Name: "go", DisplayName: "Go", StartCommand: "go run .", BuildCommand: "go build -o /dev/null", Confidence: "high"}
	case "python":
		return RuntimeResult{Name: "python", DisplayName: "Python", StartCommand: "python main.py", BuildCommand: "pip install -r requirements.txt", Confidence: "high"}
	case "ruby":
		return RuntimeResult{Name: "ruby", DisplayName: "Ruby", StartCommand: "bundle exec ruby main.rb", BuildCommand: "bundle install", Confidence: "high"}
	default:
		return RuntimeResult{Name: name, DisplayName: name, StartCommand: "", Confidence: "low"}
	}
}

func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

// LookupStartCommand returns the start command for a given runtime name.
func LookupStartCommand(runtimeName string) string {
	switch runtimeName {
	case "node":
		return "npm start"
	case "bun":
		return "bun run start"
	case "go":
		return "go run ."
	case "python":
		return "python main.py"
	case "php":
		return "php artisan serve"
	case "ruby":
		return "bundle exec ruby main.rb"
	default:
		return ""
	}
}
