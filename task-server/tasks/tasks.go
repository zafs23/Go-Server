package tasks

import "time"

type TaskRequest struct {
	Command []string      `json:"command"`
	Timeout time.Duration `json:"timeout"`
}

type TaskResult struct {
	Command    []string `json:"command"`
	ExecutedAt int64    `json:"executed_at"`
	DurationMs float64  `json:"duraton_ms"`
	ExitCode   int      `json:"exit_code"`
	Output     string   `json:"output"`
	Error      string   `json:"error"`
}
