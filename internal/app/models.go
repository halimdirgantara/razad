package app

// CreateAppRequest is the payload for creating a new app.
type CreateAppRequest struct {
	Name     string `json:"name"`
	ProjectID string `json:"project_id"`
	GitURL   string `json:"git_url,omitempty"`
	Runtime  string `json:"runtime,omitempty"`
	StartCmd string `json:"start_cmd,omitempty"`
}

// EnvVarInput is a key-value pair for environment variables.
type EnvVarInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
