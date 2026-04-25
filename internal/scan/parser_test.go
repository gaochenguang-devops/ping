package scan

import (
	"context"
	"os/exec"
	"testing"
)

func TestClassifyResultTreatsDeadlineExceededAsUnreachable(t *testing.T) {
	t.Parallel()

	status, message := classifyResult(pingMetrics{}, "", context.DeadlineExceeded)
	if status != StatusUnreachable || message != "请求超时" {
		t.Fatalf("unexpected classification: status=%s message=%s", status, message)
	}
}

func TestClassifyResultTreatsMissingPingBinaryAsError(t *testing.T) {
	t.Parallel()

	status, message := classifyResult(pingMetrics{}, "", exec.ErrNotFound)
	if status != StatusError || message != "未找到 ping 命令" {
		t.Fatalf("unexpected classification: status=%s message=%s", status, message)
	}
}
