package app

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	defaultListenHost = ""
	defaultListenPort = 8080
)

type Config struct {
	Addr        string
	OpenBrowser bool
}

func loadConfig() (Config, error) {
	return loadConfigFrom(os.Args[1:], os.LookupEnv)
}

func loadConfigFrom(args []string, lookupEnv func(string) (string, bool)) (Config, error) {
	flagSet := flag.NewFlagSet("pingtool", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	addrFlag := flagSet.String("addr", "", "listen address, for example :8080 or 0.0.0.0:8080")
	hostFlag := flagSet.String("host", "", "listen host, for example 0.0.0.0 or 127.0.0.1")
	portFlag := flagSet.Int("port", 0, "listen port, for example 8080")

	if err := flagSet.Parse(args); err != nil {
		return Config{}, err
	}

	host := defaultListenHost
	port := defaultListenPort

	if value, ok := lookupEnv("PINGTOOL_ADDR"); ok && strings.TrimSpace(value) != "" {
		parsedHost, parsedPort, err := splitListenAddr(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid PINGTOOL_ADDR: %w", err)
		}
		host = parsedHost
		port = parsedPort
	}

	if value, ok := lookupEnv("PINGTOOL_HOST"); ok && strings.TrimSpace(value) != "" {
		host = strings.TrimSpace(value)
	}

	if value, ok := lookupEnv("PINGTOOL_PORT"); ok && strings.TrimSpace(value) != "" {
		parsedPort, err := parsePort(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid PINGTOOL_PORT: %w", err)
		}
		port = parsedPort
	}

	if strings.TrimSpace(*addrFlag) != "" {
		parsedHost, parsedPort, err := splitListenAddr(*addrFlag)
		if err != nil {
			return Config{}, fmt.Errorf("invalid -addr: %w", err)
		}
		host = parsedHost
		port = parsedPort
	}

	if strings.TrimSpace(*hostFlag) != "" {
		host = strings.TrimSpace(*hostFlag)
	}

	if *portFlag != 0 {
		if _, err := normalizePort(*portFlag); err != nil {
			return Config{}, fmt.Errorf("invalid -port: %w", err)
		}
		port = *portFlag
	}

	addr, err := joinListenAddr(host, port)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Addr:        addr,
		OpenBrowser: parseBoolEnvFrom(lookupEnv, "PINGTOOL_OPEN_BROWSER", true),
	}, nil
}

func splitListenAddr(addr string) (string, int, error) {
	value := strings.TrimSpace(addr)
	if value == "" {
		return "", 0, fmt.Errorf("empty address")
	}

	if strings.HasPrefix(value, ":") {
		port, err := parsePort(strings.TrimPrefix(value, ":"))
		if err != nil {
			return "", 0, err
		}
		return "", port, nil
	}

	if port, err := parsePort(value); err == nil {
		return "", port, nil
	}

	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		return "", 0, fmt.Errorf("expected host:port")
	}

	port, err := parsePort(portText)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func joinListenAddr(host string, port int) (string, error) {
	normalizedPort, err := normalizePort(port)
	if err != nil {
		return "", err
	}

	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		return ":" + strconv.Itoa(normalizedPort), nil
	}
	return net.JoinHostPort(trimmedHost, strconv.Itoa(normalizedPort)), nil
}

func parsePort(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("port must be a number")
	}
	return normalizePort(port)
}

func normalizePort(port int) (int, error) {
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return port, nil
}

func (c Config) URL() string {
	addr := c.Addr
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}

	host := addr
	if strings.HasPrefix(host, "0.0.0.0") {
		host = "localhost" + host[len("0.0.0.0"):]
	}
	if strings.HasPrefix(host, "[::]") {
		host = "localhost" + host[len("[::]"):]
	}

	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "http://" + host
}

func parseBoolEnvFrom(lookupEnv func(string) (string, bool), name string, fallback bool) bool {
	value, ok := lookupEnv(name)
	if !ok {
		return fallback
	}

	switch strings.TrimSpace(strings.ToLower(value)) {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
