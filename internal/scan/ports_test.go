package scan

import "testing"

func TestParsePortsExpandsRangesAndDeduplicates(t *testing.T) {
	t.Parallel()

	ports, err := parsePorts("80,443 8080-8082 443", 16)
	if err != nil {
		t.Fatalf("parsePorts returned error: %v", err)
	}

	want := []int{80, 443, 8080, 8081, 8082}
	if len(ports) != len(want) {
		t.Fatalf("unexpected port count: got %d want %d", len(ports), len(want))
	}

	for i := range want {
		if ports[i] != want[i] {
			t.Fatalf("unexpected port at index %d: got %d want %d", i, ports[i], want[i])
		}
	}
}

func TestParsePortsRejectsDescendingRange(t *testing.T) {
	t.Parallel()

	_, err := parsePorts("100-80", 16)
	if err == nil {
		t.Fatal("expected error for descending port range")
	}
}

func TestBuildScanTasksForBothMode(t *testing.T) {
	t.Parallel()

	targets, err := expandTargets("127.0.0.1 127.0.0.2", "", 16)
	if err != nil {
		t.Fatalf("expandTargets returned error: %v", err)
	}

	tasks, err := buildScanTasks(targets, ScanModeBoth, []int{80, 443}, 16)
	if err != nil {
		t.Fatalf("buildScanTasks returned error: %v", err)
	}

	if len(tasks) != 6 {
		t.Fatalf("unexpected task count: got %d want %d", len(tasks), 6)
	}

	if tasks[0].Kind != ScanKindPing || tasks[0].Address != "127.0.0.1" {
		t.Fatalf("unexpected first task: %+v", tasks[0])
	}
	if tasks[1].Kind != ScanKindTCP || tasks[1].Port != 80 {
		t.Fatalf("unexpected second task: %+v", tasks[1])
	}
	if tasks[2].Kind != ScanKindTCP || tasks[2].Port != 443 {
		t.Fatalf("unexpected third task: %+v", tasks[2])
	}
}
