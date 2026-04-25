package scan

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestScanTCPReachable(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	service := NewService()
	port := listener.Addr().(*net.TCPAddr).Port
	result := service.scanTCP(context.Background(), scanTask{
		Address: "127.0.0.1",
		Source:  SourceManual,
		Kind:    ScanKindTCP,
		Port:    port,
	}, normalizedRequest{
		Timeout: time.Second,
	})

	<-accepted

	if result.Status != StatusReachable {
		t.Fatalf("unexpected status: got %s message=%s", result.Status, result.Message)
	}
	if result.Endpoint == "" || result.Port == nil || *result.Port != port {
		t.Fatalf("unexpected endpoint fields: %+v", result)
	}
}

func TestClassifyDialErrorTimeout(t *testing.T) {
	t.Parallel()

	status, message := classifyDialError(context.DeadlineExceeded)
	if status != StatusUnreachable || message == "" {
		t.Fatalf("unexpected classification: status=%s message=%s", status, message)
	}
}

func TestClassifyDialErrorRefused(t *testing.T) {
	t.Parallel()

	status, message := classifyDialError(errors.New("connectex: No connection could be made because the target machine actively refused it"))
	if status != StatusUnreachable || message != "端口关闭" {
		t.Fatalf("unexpected classification: status=%s message=%s", status, message)
	}
}
