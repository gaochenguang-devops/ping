package scan

import (
	"fmt"
	"regexp"
	"strconv"
)

var portRangePattern = regexp.MustCompile(`^(\d+)-(\d+)$`)

func parsePorts(input string, limit int) ([]int, error) {
	entries := splitEntries(input)
	if len(entries) == 0 {
		return nil, nil
	}

	ports := make([]int, 0, len(entries))
	seen := make(map[int]struct{}, len(entries))

	addPort := func(port int) error {
		if port < 1 || port > 65535 {
			return ValidationError{Message: fmt.Sprintf("无效端口: %d", port)}
		}
		if _, ok := seen[port]; ok {
			return nil
		}
		if len(ports) >= limit {
			return ValidationError{Message: fmt.Sprintf("端口数量不能超过 %d 个", limit)}
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
		return nil
	}

	for _, entry := range entries {
		if match := portRangePattern.FindStringSubmatch(entry); len(match) == 3 {
			start, _ := strconv.Atoi(match[1])
			end, _ := strconv.Atoi(match[2])
			if start > end {
				return nil, ValidationError{Message: fmt.Sprintf("端口范围起始值不能大于结束值: %s", entry)}
			}
			for port := start; port <= end; port++ {
				if err := addPort(port); err != nil {
					return nil, err
				}
			}
			continue
		}

		port, err := strconv.Atoi(entry)
		if err != nil {
			return nil, ValidationError{Message: fmt.Sprintf("无效端口: %s", entry)}
		}
		if err := addPort(port); err != nil {
			return nil, err
		}
	}

	return ports, nil
}
