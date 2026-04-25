package scan

import "testing"

func TestExpandTargetsShortIPv4Range(t *testing.T) {
	t.Parallel()

	targets, err := expandTargets("192.168.1.1-3", "", 16)
	if err != nil {
		t.Fatalf("expandTargets returned error: %v", err)
	}

	got := make([]string, 0, len(targets))
	for _, target := range targets {
		got = append(got, target.Address)
	}

	want := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected target count: got %d want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected target at index %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestExpandTargetsFullIPv4Range(t *testing.T) {
	t.Parallel()

	targets, err := expandTargets("192.168.1.9-192.168.1.11", "", 16)
	if err != nil {
		t.Fatalf("expandTargets returned error: %v", err)
	}

	got := make([]string, 0, len(targets))
	for _, target := range targets {
		got = append(got, target.Address)
	}

	want := []string{
		"192.168.1.9",
		"192.168.1.10",
		"192.168.1.11",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected target count: got %d want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected target at index %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestExpandTargetsFullIPv6Range(t *testing.T) {
	t.Parallel()

	targets, err := expandTargets("2001:db8::1-2001:db8::3", "", 16)
	if err != nil {
		t.Fatalf("expandTargets returned error: %v", err)
	}

	got := make([]string, 0, len(targets))
	for _, target := range targets {
		got = append(got, target.Address)
	}

	want := []string{
		"2001:db8::1",
		"2001:db8::2",
		"2001:db8::3",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected target count: got %d want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected target at index %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestExpandTargetsIPv6CIDR(t *testing.T) {
	t.Parallel()

	targets, err := expandTargets("", "2001:db8::/126", 16)
	if err != nil {
		t.Fatalf("expandTargets returned error: %v", err)
	}

	got := make([]string, 0, len(targets))
	for _, target := range targets {
		got = append(got, target.Address)
	}

	want := []string{
		"2001:db8::",
		"2001:db8::1",
		"2001:db8::2",
		"2001:db8::3",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected target count: got %d want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected target at index %d: got %s want %s", i, got[i], want[i])
		}
	}
}

func TestExpandTargetsRejectsDescendingRange(t *testing.T) {
	t.Parallel()

	_, err := expandTargets("192.168.1.10-1", "", 16)
	if err == nil {
		t.Fatal("expected error for descending range")
	}
}

func TestExpandTargetsRejectsOversizedRange(t *testing.T) {
	t.Parallel()

	_, err := expandTargets("192.168.1.1-192.168.1.20", "", 8)
	if err == nil {
		t.Fatal("expected error for oversized range")
	}
}
