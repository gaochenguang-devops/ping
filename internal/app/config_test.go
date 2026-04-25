package app

import "testing"

func TestLoadConfigFromDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFrom(nil, func(string) (string, bool) {
		return "", false
	})
	if err != nil {
		t.Fatalf("loadConfigFrom returned error: %v", err)
	}

	if cfg.Addr != ":8080" {
		t.Fatalf("unexpected addr: got %s want %s", cfg.Addr, ":8080")
	}
}

func TestLoadConfigFromEnvPort(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFrom(nil, func(name string) (string, bool) {
		switch name {
		case "PINGTOOL_PORT":
			return "9090", true
		default:
			return "", false
		}
	})
	if err != nil {
		t.Fatalf("loadConfigFrom returned error: %v", err)
	}

	if cfg.Addr != ":9090" {
		t.Fatalf("unexpected addr: got %s want %s", cfg.Addr, ":9090")
	}
}

func TestLoadConfigFromFlagPortOverridesEnv(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFrom([]string{"-port", "7070"}, func(name string) (string, bool) {
		switch name {
		case "PINGTOOL_PORT":
			return "9090", true
		default:
			return "", false
		}
	})
	if err != nil {
		t.Fatalf("loadConfigFrom returned error: %v", err)
	}

	if cfg.Addr != ":7070" {
		t.Fatalf("unexpected addr: got %s want %s", cfg.Addr, ":7070")
	}
}

func TestLoadConfigFromAddrAndPortFlag(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFrom([]string{"-addr", "127.0.0.1:8088", "-port", "8099"}, func(string) (string, bool) {
		return "", false
	})
	if err != nil {
		t.Fatalf("loadConfigFrom returned error: %v", err)
	}

	if cfg.Addr != "127.0.0.1:8099" {
		t.Fatalf("unexpected addr: got %s want %s", cfg.Addr, "127.0.0.1:8099")
	}
}

func TestLoadConfigFromRejectsInvalidPort(t *testing.T) {
	t.Parallel()

	_, err := loadConfigFrom([]string{"-port", "70000"}, func(string) (string, bool) {
		return "", false
	})
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
