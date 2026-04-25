package scan

import (
	"context"
	"math"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func (s *Service) scanTarget(ctx context.Context, task scanTask, req normalizedRequest) ScanResult {
	switch task.Kind {
	case ScanKindTCP:
		return s.scanTCP(ctx, task, req)
	default:
		return s.scanPing(ctx, task, req)
	}
}

func (s *Service) scanPing(ctx context.Context, task scanTask, req normalizedRequest) ScanResult {
	startedAt := time.Now()

	result := ScanResult{
		Target:  task.Address,
		Source:  task.Source,
		Kind:    ScanKindPing,
		Message: "-",
	}

	defer func() {
		result.DurationMS = time.Since(startedAt).Milliseconds()
	}()

	if ip := net.ParseIP(task.Address); ip != nil {
		result.ResolvedIPs = []string{ip.String()}
	}

	if req.ResolveDNS && net.ParseIP(task.Address) == nil {
		ips, err := resolveHost(ctx, task.Address)
		if err != nil {
			result.Status = StatusError
			result.Message = "DNS 解析失败: " + err.Error()
			return result
		}
		result.ResolvedIPs = ips
	}

	output, execErr := runPing(ctx, task.Address, req)
	metrics := parsePingOutput(output)

	result.Sent = metrics.Sent
	result.Received = metrics.Received
	result.LossPercent = metrics.LossPercent
	result.MinLatencyMS = metrics.MinLatencyMS
	result.AvgLatencyMS = metrics.AvgLatencyMS
	result.MaxLatencyMS = metrics.MaxLatencyMS

	result.Status, result.Message = classifyResult(metrics, output, execErr)
	result.Reachable = result.Status == StatusReachable

	return result
}

func resolveHost(ctx context.Context, host string) ([]string, error) {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(ips))
	values := make([]string, 0, len(ips))
	for _, ip := range ips {
		text := ip.String()
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		values = append(values, text)
	}
	return values, nil
}

func runPing(parent context.Context, target string, req normalizedRequest) (string, error) {
	timeout := time.Duration(req.Count)*req.Timeout + 3*time.Second
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := buildPingCommand(ctx, target, req)
	output, err := cmd.CombinedOutput()
	return decodePingOutput(output), err
}

func buildPingCommand(ctx context.Context, target string, req normalizedRequest) *exec.Cmd {
	name, args := pingCommandSpec(target, req, runtimeGOOS())
	return exec.CommandContext(ctx, name, args...)
}

func pingCommandSpec(target string, req normalizedRequest, goos string) (string, []string) {
	forceIPv6 := isIPv6Literal(target)

	switch goos {
	case "windows":
		args := make([]string, 0, 6)
		if forceIPv6 {
			args = append(args, "-6")
		}
		args = append(args,
			"-n", strconv.Itoa(req.Count),
			"-w", strconv.Itoa(req.TimeoutMS),
			target,
		)
		return "ping", args
	case "linux":
		waitSeconds := strconv.Itoa(int(math.Max(1, math.Ceil(float64(req.TimeoutMS)/1000.0))))
		args := make([]string, 0, 8)
		if forceIPv6 {
			args = append(args, "-6")
		}
		args = append(args,
			"-c", strconv.Itoa(req.Count),
			"-i", "0.2",
			"-W", waitSeconds,
			target,
		)
		return "ping", args
	case "darwin":
		args := make([]string, 0, 6)
		if forceIPv6 {
			args = append(args, "-6")
		}
		args = append(args,
			"-c", strconv.Itoa(req.Count),
			target,
		)
		return "ping", args
	default:
		args := make([]string, 0, 6)
		if forceIPv6 {
			args = append(args, "-6")
		}
		args = append(args,
			"-c", strconv.Itoa(req.Count),
			target,
		)
		return "ping", args
	}
}

func decodePingOutput(output []byte) string {
	if runtime.GOOS != "windows" {
		return strings.TrimSpace(string(output))
	}

	decoded, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), string(output))
	if err != nil {
		return strings.TrimSpace(string(output))
	}

	return strings.TrimSpace(decoded)
}

func isIPv6Literal(target string) bool {
	ip := net.ParseIP(target)
	return ip != nil && ip.To4() == nil
}

func runtimeGOOS() string {
	return runtimeOS()
}

var runtimeOS = func() string {
	return runtime.GOOS
}
