package scan

import (
	"context"
	"strings"
	"sync"
	"time"
)

const (
	timeLayout = "2006-01-02 15:04:05"

	defaultCount       = 1
	defaultTimeoutMS   = 1000
	defaultConcurrency = 32
	defaultMode        = ScanModePing

	minCount       = 1
	maxCount       = 4
	minTimeoutMS   = 200
	maxTimeoutMS   = 5000
	minConcurrency = 1
	maxConcurrency = 256
	maxTargets     = 4096
	maxPorts       = 256
	maxScanTasks   = 65536
)

type Service struct{}

type normalizedRequest struct {
	Count       int
	Timeout     time.Duration
	TimeoutMS   int
	Concurrency int
	ResolveDNS  bool
	Mode        string
	Ports       []int
}

type scanTask struct {
	Address string
	Source  string
	Kind    string
	Port    int
	Index   int
}

type indexedResult struct {
	Index  int
	Result ScanResult
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Scan(ctx context.Context, req ScanRequest) (ScanResponse, error) {
	return s.ScanWithProgress(ctx, req, nil)
}

func (s *Service) ScanWithProgress(ctx context.Context, req ScanRequest, report ProgressFunc) (ScanResponse, error) {
	options, err := normalizeRequest(req)
	if err != nil {
		return ScanResponse{}, err
	}

	targets, err := expandTargets(req.Targets, req.CIDR, maxTargets)
	if err != nil {
		return ScanResponse{}, err
	}

	tasks, err := buildScanTasks(targets, options.Mode, options.Ports, maxScanTasks)
	if err != nil {
		return ScanResponse{}, err
	}

	startedAt := time.Now()
	progress := ScanProgress{
		Total:     len(tasks),
		Status:    ProgressStatusRunning,
		StartedAt: startedAt.Format(timeLayout),
		Message:   "检测进行中",
	}
	emitProgress(report, progress)

	workerCount := minInt(options.Concurrency, len(tasks))

	jobs := make(chan scanTask)
	results := make(chan indexedResult, len(tasks))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				result := s.scanTarget(ctx, task, options)

				select {
				case <-ctx.Done():
					return
				case results <- indexedResult{Index: task.Index, Result: result}:
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case jobs <- task:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	ordered := make([]ScanResult, len(tasks))
	completed := 0
	reachable := 0
	unreachable := 0
	errorsCount := 0

	for completed < len(tasks) {
		select {
		case <-ctx.Done():
			progress.Completed = completed
			progress.Reachable = reachable
			progress.Unreachable = unreachable
			progress.Errors = errorsCount
			progress.Percent = calculatePercent(completed, len(tasks))
			progress.Status = ProgressStatusCanceled
			progress.FinishedAt = time.Now().Format(timeLayout)
			progress.Message = "检测已取消"
			emitProgress(report, progress)
			return ScanResponse{}, ctx.Err()
		case item, ok := <-results:
			if !ok {
				completed = len(tasks)
				break
			}
			ordered[item.Index] = item.Result
			completed++

			switch item.Result.Status {
			case StatusReachable:
				reachable++
			case StatusUnreachable:
				unreachable++
			default:
				errorsCount++
			}

			progress.Completed = completed
			progress.Reachable = reachable
			progress.Unreachable = unreachable
			progress.Errors = errorsCount
			progress.Percent = calculatePercent(completed, len(tasks))
			progress.Status = ProgressStatusRunning
			progress.Message = "检测进行中"
			emitProgress(report, progress)
		}
	}

	finishedAt := time.Now()
	progress.Completed = len(tasks)
	progress.Reachable = reachable
	progress.Unreachable = unreachable
	progress.Errors = errorsCount
	progress.Percent = 100
	progress.Status = ProgressStatusDone
	progress.FinishedAt = finishedAt.Format(timeLayout)
	progress.Message = "检测完成"
	emitProgress(report, progress)

	return ScanResponse{
		Summary: buildSummary(ordered, startedAt, finishedAt),
		Results: ordered,
	}, nil
}

func normalizeRequest(req ScanRequest) (normalizedRequest, error) {
	count := req.Count
	if count == 0 {
		count = defaultCount
	}
	if count < minCount || count > maxCount {
		return normalizedRequest{}, ValidationError{Message: "请求次数必须在 1 到 4 之间"}
	}

	timeoutMS := req.TimeoutMS
	if timeoutMS == 0 {
		timeoutMS = defaultTimeoutMS
	}
	if timeoutMS < minTimeoutMS || timeoutMS > maxTimeoutMS {
		return normalizedRequest{}, ValidationError{Message: "超时时间必须在 200 到 5000 毫秒之间"}
	}

	concurrency := req.Concurrency
	if concurrency == 0 {
		concurrency = defaultConcurrency
	}
	if concurrency < minConcurrency || concurrency > maxConcurrency {
		return normalizedRequest{}, ValidationError{Message: "并发数必须在 1 到 256 之间"}
	}

	resolveDNS := true
	if req.ResolveDNS != nil {
		resolveDNS = *req.ResolveDNS
	}

	mode := strings.TrimSpace(strings.ToLower(req.Mode))
	if mode == "" {
		mode = defaultMode
	}
	switch mode {
	case ScanModePing, ScanModeTCP, ScanModeBoth:
	default:
		return normalizedRequest{}, ValidationError{Message: "检测模式只支持 ping、tcp 或 both"}
	}

	ports, err := parsePorts(req.Ports, maxPorts)
	if err != nil {
		return normalizedRequest{}, err
	}
	if mode == ScanModePing && len(ports) > 0 {
		return normalizedRequest{}, ValidationError{Message: "Ping 模式下不需要输入端口"}
	}
	if mode != ScanModePing && len(ports) == 0 {
		return normalizedRequest{}, ValidationError{Message: "TCP 端口探测需要输入要检测的端口"}
	}

	return normalizedRequest{
		Count:       count,
		Timeout:     time.Duration(timeoutMS) * time.Millisecond,
		TimeoutMS:   timeoutMS,
		Concurrency: concurrency,
		ResolveDNS:  resolveDNS,
		Mode:        mode,
		Ports:       ports,
	}, nil
}

func buildScanTasks(targets []targetSpec, mode string, ports []int, limit int) ([]scanTask, error) {
	taskCount := len(targets)
	switch mode {
	case ScanModeTCP:
		taskCount = len(targets) * len(ports)
	case ScanModeBoth:
		taskCount = len(targets) * (len(ports) + 1)
	}

	if taskCount > limit {
		return nil, ValidationError{Message: "检测任务数量过多，请收窄目标或端口范围"}
	}

	tasks := make([]scanTask, 0, taskCount)
	index := 0
	for _, target := range targets {
		if mode == ScanModePing || mode == ScanModeBoth {
			tasks = append(tasks, scanTask{
				Address: target.Address,
				Source:  target.Source,
				Kind:    ScanKindPing,
				Index:   index,
			})
			index++
		}

		if mode == ScanModeTCP || mode == ScanModeBoth {
			for _, port := range ports {
				tasks = append(tasks, scanTask{
					Address: target.Address,
					Source:  target.Source,
					Kind:    ScanKindTCP,
					Port:    port,
					Index:   index,
				})
				index++
			}
		}
	}

	return tasks, nil
}

func buildSummary(results []ScanResult, startedAt, finishedAt time.Time) ScanSummary {
	summary := ScanSummary{
		Total:      len(results),
		ElapsedMS:  finishedAt.Sub(startedAt).Milliseconds(),
		StartedAt:  startedAt.Format(timeLayout),
		FinishedAt: finishedAt.Format(timeLayout),
	}

	var latencySum float64
	var latencyCount int

	for _, result := range results {
		switch result.Status {
		case StatusReachable:
			summary.Reachable++
		case StatusUnreachable:
			summary.Unreachable++
		default:
			summary.Errors++
		}

		if result.AvgLatencyMS != nil {
			latencySum += *result.AvgLatencyMS
			latencyCount++
		}
	}

	if latencyCount > 0 {
		avg := latencySum / float64(latencyCount)
		summary.AvgLatencyMS = &avg
	}

	return summary
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func calculatePercent(completed, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(completed) * 100 / float64(total)
}

func emitProgress(report ProgressFunc, progress ScanProgress) {
	if report != nil {
		report(progress)
	}
}
