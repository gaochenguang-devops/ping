package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"pingtool/internal/scan"
)

const jobTimeLayout = "2006-01-02 15:04:05"

type scanJobManager struct {
	scanner *scan.Service

	mu   sync.RWMutex
	jobs map[string]*scanJob
}

type scanJob struct {
	ID       string
	Progress scan.ScanProgress
	Result   *scan.ScanResponse
	Error    string

	cancel context.CancelFunc
}

type scanJobSnapshot struct {
	ID       string             `json:"id"`
	Progress scan.ScanProgress  `json:"progress"`
	Result   *scan.ScanResponse `json:"result,omitempty"`
	Error    string             `json:"error,omitempty"`
}

func newScanJobManager(scanner *scan.Service) *scanJobManager {
	return &scanJobManager{
		scanner: scanner,
		jobs:    make(map[string]*scanJob),
	}
}

func (m *scanJobManager) Create(req scan.ScanRequest) (scanJobSnapshot, error) {
	id, err := newJobID()
	if err != nil {
		return scanJobSnapshot{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	job := &scanJob{
		ID: id,
		Progress: scan.ScanProgress{
			Status:  scan.ProgressStatusQueued,
			Message: "任务已创建，等待开始",
		},
		cancel: cancel,
	}

	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	go m.run(job, ctx, req)

	return m.snapshot(id)
}

func (m *scanJobManager) Snapshot(id string) (scanJobSnapshot, bool) {
	snapshot, err := m.snapshot(id)
	if err != nil {
		return scanJobSnapshot{}, false
	}
	return snapshot, true
}

func (m *scanJobManager) Cancel(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	job := m.jobs[id]
	if job == nil {
		return false
	}

	if isTerminalJobStatus(job.Progress.Status) {
		return true
	}

	job.cancel()
	job.Progress.Status = scan.ProgressStatusCanceled
	job.Progress.FinishedAt = time.Now().Format(jobTimeLayout)
	job.Progress.Message = "取消请求已提交"
	return true
}

func (m *scanJobManager) run(job *scanJob, ctx context.Context, req scan.ScanRequest) {
	resp, err := m.scanner.ScanWithProgress(ctx, req, func(progress scan.ScanProgress) {
		m.updateProgress(job.ID, progress)
	})

	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.jobs[job.ID]
	if current == nil {
		return
	}

	if err == nil {
		current.Result = &resp
		current.Progress.Status = scan.ProgressStatusDone
		current.Progress.Percent = 100
		if current.Progress.FinishedAt == "" {
			current.Progress.FinishedAt = time.Now().Format(jobTimeLayout)
		}
		current.Progress.Message = "检测完成"
		return
	}

	if errors.Is(err, context.Canceled) {
		current.Progress.Status = scan.ProgressStatusCanceled
		if current.Progress.FinishedAt == "" {
			current.Progress.FinishedAt = time.Now().Format(jobTimeLayout)
		}
		current.Progress.Message = "检测已取消"
		return
	}

	current.Error = err.Error()
	current.Progress.Status = scan.ProgressStatusError
	current.Progress.Message = err.Error()
	if current.Progress.FinishedAt == "" {
		current.Progress.FinishedAt = time.Now().Format(jobTimeLayout)
	}
}

func (m *scanJobManager) updateProgress(id string, progress scan.ScanProgress) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job := m.jobs[id]
	if job == nil {
		return
	}

	if isTerminalJobStatus(job.Progress.Status) && !isTerminalJobStatus(progress.Status) {
		return
	}

	job.Progress = progress
}

func (m *scanJobManager) snapshot(id string) (scanJobSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job := m.jobs[id]
	if job == nil {
		return scanJobSnapshot{}, errors.New("job not found")
	}

	var result *scan.ScanResponse
	if job.Result != nil {
		copyValue := *job.Result
		result = &copyValue
	}

	return scanJobSnapshot{
		ID:       job.ID,
		Progress: job.Progress,
		Result:   result,
		Error:    job.Error,
	}, nil
}

func isTerminalJobStatus(status string) bool {
	switch status {
	case scan.ProgressStatusDone, scan.ProgressStatusError, scan.ProgressStatusCanceled:
		return true
	default:
		return false
	}
}

func newJobID() (string, error) {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
