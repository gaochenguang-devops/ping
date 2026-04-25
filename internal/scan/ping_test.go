package scan

import (
	"slices"
	"testing"
	"time"
)

func TestPingCommandSpecAddsIPv6FlagForIPv6Literal(t *testing.T) {
	t.Parallel()

	_, args := pingCommandSpec("::1", normalizedRequest{
		Count:     1,
		Timeout:   time.Second,
		TimeoutMS: 1000,
	}, "linux")

	if !slices.Contains(args, "-6") {
		t.Fatalf("expected -6 flag in command args, got %v", args)
	}
}

func TestPingCommandSpecDoesNotAddIPv6FlagForIPv4Literal(t *testing.T) {
	t.Parallel()

	_, args := pingCommandSpec("127.0.0.1", normalizedRequest{
		Count:     1,
		Timeout:   time.Second,
		TimeoutMS: 1000,
	}, "linux")

	if slices.Contains(args, "-6") {
		t.Fatalf("did not expect -6 flag in command args, got %v", args)
	}
}

func TestPingCommandSpecForWindows(t *testing.T) {
	t.Parallel()

	name, args := pingCommandSpec("127.0.0.1", normalizedRequest{
		Count:     2,
		Timeout:   time.Second,
		TimeoutMS: 800,
	}, "windows")

	if name != "ping" {
		t.Fatalf("unexpected command name: %s", name)
	}
	if !slices.Contains(args, "-n") || !slices.Contains(args, "-w") {
		t.Fatalf("expected windows flags in args, got %v", args)
	}
}

func TestPingCommandSpecForLinux(t *testing.T) {
	t.Parallel()

	_, args := pingCommandSpec("127.0.0.1", normalizedRequest{
		Count:     2,
		Timeout:   time.Second,
		TimeoutMS: 800,
	}, "linux")

	if !slices.Contains(args, "-W") || !slices.Contains(args, "-i") {
		t.Fatalf("expected linux flags in args, got %v", args)
	}
}

func TestPingCommandSpecForDarwinAvoidsLinuxSpecificTimeoutFlags(t *testing.T) {
	t.Parallel()

	_, args := pingCommandSpec("127.0.0.1", normalizedRequest{
		Count:     2,
		Timeout:   time.Second,
		TimeoutMS: 800,
	}, "darwin")

	if slices.Contains(args, "-W") || slices.Contains(args, "-i") {
		t.Fatalf("did not expect linux-specific flags in darwin args, got %v", args)
	}
	if !slices.Contains(args, "-c") {
		t.Fatalf("expected count flag in darwin args, got %v", args)
	}
}
