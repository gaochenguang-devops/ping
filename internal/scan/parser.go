package scan

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	unixPacketsPattern    = regexp.MustCompile(`(?im)(\d+)\s+packets transmitted,\s+(\d+)\s+(?:packets )?received,\s+(\d+(?:\.\d+)?)%\s+packet loss`)
	sentPattern           = regexp.MustCompile(`(?im)(?:sent|已发送)\s*=\s*(\d+)`)
	receivedPattern       = regexp.MustCompile(`(?im)(?:received|已接收)\s*=\s*(\d+)`)
	lossPattern           = regexp.MustCompile(`(?im)\((\d+(?:\.\d+)?)%\s*(?:loss|丢失)\)`)
	unixLatencyPattern    = regexp.MustCompile(`(?im)(?:rtt|round-trip)[^=]*=\s*(\d+(?:\.\d+)?)\/(\d+(?:\.\d+)?)\/(\d+(?:\.\d+)?)(?:\/\d+(?:\.\d+)?)?\s*ms`)
	windowsLatencyPattern = regexp.MustCompile(`(?im)(?:minimum|最短)\s*=\s*(\d+(?:\.\d+)?)ms[，,]\s*(?:maximum|最长)\s*=\s*(\d+(?:\.\d+)?)ms[，,]\s*(?:average|平均)\s*=\s*(\d+(?:\.\d+)?)ms`)
	replyPattern          = regexp.MustCompile(`(?im)(ttl[=\s]|bytes from|reply from|time[=<]?\s*\d|时间[=<]?\s*\d|来自 .* 的回复)`)
)

var fatalMessages = map[string]string{
	"could not find host":                  "主机名无法解析",
	"ping request could not find host":     "主机名无法解析",
	"请求找不到主机":                              "主机名无法解析",
	"name or service not known":            "主机名无法解析",
	"unknown host":                         "主机名无法解析",
	"temporary failure in name resolution": "DNS 解析失败",
	"general failure":                      "网络返回一般故障",
	"transmit failed":                      "发送失败",
}

var unreachableMarkers = []string{
	"request timed out",
	"请求超时",
	"destination host unreachable",
	"destination net unreachable",
	"无法访问目标主机",
	"100% packet loss",
	"100% 丢失",
}

type pingMetrics struct {
	Sent         int
	Received     int
	LossPercent  float64
	HasReply     bool
	MinLatencyMS *float64
	AvgLatencyMS *float64
	MaxLatencyMS *float64
}

func parsePingOutput(output string) pingMetrics {
	normalized := strings.ReplaceAll(output, "\r\n", "\n")

	metrics := pingMetrics{
		HasReply: replyPattern.MatchString(normalized),
	}

	if match := unixPacketsPattern.FindStringSubmatch(normalized); len(match) == 4 {
		metrics.Sent = mustInt(match[1])
		metrics.Received = mustInt(match[2])
		metrics.LossPercent = mustFloat(match[3])
	} else {
		if match := sentPattern.FindStringSubmatch(normalized); len(match) == 2 {
			metrics.Sent = mustInt(match[1])
		}
		if match := receivedPattern.FindStringSubmatch(normalized); len(match) == 2 {
			metrics.Received = mustInt(match[1])
		}
		if match := lossPattern.FindStringSubmatch(normalized); len(match) == 2 {
			metrics.LossPercent = mustFloat(match[1])
		}
	}

	if metrics.Sent > 0 && metrics.LossPercent == 0 {
		metrics.LossPercent = float64(metrics.Sent-metrics.Received) * 100 / float64(metrics.Sent)
	}

	if match := unixLatencyPattern.FindStringSubmatch(normalized); len(match) >= 4 {
		metrics.MinLatencyMS = floatPtr(match[1])
		metrics.AvgLatencyMS = floatPtr(match[2])
		metrics.MaxLatencyMS = floatPtr(match[3])
	} else if match := windowsLatencyPattern.FindStringSubmatch(normalized); len(match) == 4 {
		metrics.MinLatencyMS = floatPtr(match[1])
		metrics.MaxLatencyMS = floatPtr(match[2])
		metrics.AvgLatencyMS = floatPtr(match[3])
	}

	return metrics
}

func classifyResult(metrics pingMetrics, output string, execErr error) (string, string) {
	lower := strings.ToLower(output)

	if errors.Is(execErr, exec.ErrNotFound) {
		return StatusError, "未找到 ping 命令"
	}

	for needle, message := range fatalMessages {
		if strings.Contains(lower, needle) {
			return StatusError, message
		}
	}

	if metrics.Received > 0 || metrics.HasReply {
		if metrics.AvgLatencyMS != nil {
			return StatusReachable, fmt.Sprintf("平均 %.2f ms", *metrics.AvgLatencyMS)
		}
		return StatusReachable, "可达"
	}

	if errors.Is(execErr, context.DeadlineExceeded) {
		return StatusUnreachable, "请求超时"
	}

	if metrics.Sent > 0 || containsAny(lower, unreachableMarkers) {
		return StatusUnreachable, "无响应或目标不可达"
	}

	if execErr != nil {
		if line := firstMeaningfulLine(output); line != "" {
			return StatusError, line
		}
		return StatusError, execErr.Error()
	}

	if line := firstMeaningfulLine(output); line != "" {
		return StatusError, line
	}

	return StatusError, "未获得有效结果"
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func firstMeaningfulLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func floatPtr(value string) *float64 {
	parsed := mustFloat(value)
	return &parsed
}

func mustInt(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func mustFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}
