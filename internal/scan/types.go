package scan

const (
	StatusReachable   = "reachable"
	StatusUnreachable = "unreachable"
	StatusError       = "error"

	SourceManual = "manual"
	SourceCIDR   = "cidr"

	ScanModePing = "ping"
	ScanModeTCP  = "tcp"
	ScanModeBoth = "both"

	ScanKindPing = "ping"
	ScanKindTCP  = "tcp"

	ProgressStatusQueued   = "queued"
	ProgressStatusRunning  = "running"
	ProgressStatusDone     = "done"
	ProgressStatusError    = "error"
	ProgressStatusCanceled = "canceled"
)

type ScanRequest struct {
	Targets     string `json:"targets"`
	CIDR        string `json:"cidr"`
	Ports       string `json:"ports"`
	Mode        string `json:"mode"`
	Count       int    `json:"count"`
	TimeoutMS   int    `json:"timeout_ms"`
	Concurrency int    `json:"concurrency"`
	ResolveDNS  *bool  `json:"resolve_dns"`
}

type ScanResponse struct {
	Summary ScanSummary  `json:"summary"`
	Results []ScanResult `json:"results"`
}

type ScanProgress struct {
	Total       int     `json:"total"`
	Completed   int     `json:"completed"`
	Reachable   int     `json:"reachable"`
	Unreachable int     `json:"unreachable"`
	Errors      int     `json:"errors"`
	Percent     float64 `json:"percent"`
	Status      string  `json:"status"`
	StartedAt   string  `json:"started_at,omitempty"`
	FinishedAt  string  `json:"finished_at,omitempty"`
	Message     string  `json:"message,omitempty"`
}

type ProgressFunc func(ScanProgress)

type ScanSummary struct {
	Total        int      `json:"total"`
	Reachable    int      `json:"reachable"`
	Unreachable  int      `json:"unreachable"`
	Errors       int      `json:"errors"`
	AvgLatencyMS *float64 `json:"avg_latency_ms,omitempty"`
	ElapsedMS    int64    `json:"elapsed_ms"`
	StartedAt    string   `json:"started_at"`
	FinishedAt   string   `json:"finished_at"`
}

type ScanResult struct {
	Target       string   `json:"target"`
	Source       string   `json:"source"`
	Kind         string   `json:"kind"`
	Endpoint     string   `json:"endpoint,omitempty"`
	Port         *int     `json:"port,omitempty"`
	Status       string   `json:"status"`
	Reachable    bool     `json:"reachable"`
	ResolvedIPs  []string `json:"resolved_ips,omitempty"`
	Sent         int      `json:"sent"`
	Received     int      `json:"received"`
	LossPercent  float64  `json:"loss_percent"`
	MinLatencyMS *float64 `json:"min_latency_ms,omitempty"`
	AvgLatencyMS *float64 `json:"avg_latency_ms,omitempty"`
	MaxLatencyMS *float64 `json:"max_latency_ms,omitempty"`
	DurationMS   int64    `json:"duration_ms"`
	Message      string   `json:"message"`
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
