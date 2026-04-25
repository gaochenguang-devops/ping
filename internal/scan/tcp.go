package scan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func (s *Service) scanTCP(parent context.Context, task scanTask, req normalizedRequest) ScanResult {
	startedAt := time.Now()
	port := task.Port
	endpoint := net.JoinHostPort(task.Address, strconv.Itoa(port))

	result := ScanResult{
		Target:   task.Address,
		Source:   task.Source,
		Kind:     ScanKindTCP,
		Endpoint: endpoint,
		Port:     &port,
		Message:  "-",
	}

	defer func() {
		result.DurationMS = time.Since(startedAt).Milliseconds()
	}()

	if ip := net.ParseIP(task.Address); ip != nil {
		result.ResolvedIPs = []string{ip.String()}
	}

	if req.ResolveDNS && net.ParseIP(task.Address) == nil {
		ips, err := resolveHost(parent, task.Address)
		if err != nil {
			result.Status = StatusError
			result.Message = "DNS 解析失败: " + err.Error()
			return result
		}
		result.ResolvedIPs = ips
	}

	ctx, cancel := context.WithTimeout(parent, req.Timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: req.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		result.Status, result.Message = classifyDialError(err)
		result.Reachable = false
		return result
	}
	defer conn.Close()

	duration := float64(time.Since(startedAt)) / float64(time.Millisecond)
	result.Status = StatusReachable
	result.Reachable = true
	result.Message = fmt.Sprintf("TCP %d 开放", port)
	result.MinLatencyMS = &duration
	result.AvgLatencyMS = &duration
	result.MaxLatencyMS = &duration

	return result
}

func classifyDialError(err error) (string, string) {
	if errors.Is(err, context.DeadlineExceeded) {
		return StatusUnreachable, "连接超时"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return StatusUnreachable, "连接超时"
	}

	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "connection refused"),
		strings.Contains(text, "actively refused"):
		return StatusUnreachable, "端口关闭"
	case strings.Contains(text, "no route to host"),
		strings.Contains(text, "network is unreachable"),
		strings.Contains(text, "host is down"),
		strings.Contains(text, "host unreachable"):
		return StatusUnreachable, "目标不可达"
	case strings.Contains(text, "no such host"),
		strings.Contains(text, "server misbehaving"),
		strings.Contains(text, "lookup "):
		return StatusError, "DNS 解析失败"
	default:
		return StatusError, firstMeaningfulLine(err.Error())
	}
}
