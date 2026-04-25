package scan

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
)

var (
	hostPattern           = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\-._:]*$`)
	splitPattern          = regexp.MustCompile(`[\s,;]+`)
	shortIPv4RangePattern = regexp.MustCompile(`^(\d{1,3}(?:\.\d{1,3}){3})-(\d{1,3})$`)
	fullIPv4RangePattern  = regexp.MustCompile(`^(\d{1,3}(?:\.\d{1,3}){3})-(\d{1,3}(?:\.\d{1,3}){3})$`)
	fullIPv6RangePattern  = regexp.MustCompile(`^([0-9A-Fa-f:]+)-([0-9A-Fa-f:]+)$`)
)

type targetSpec struct {
	Address string
	Source  string
	Index   int
}

func expandTargets(targetInput, cidrInput string, limit int) ([]targetSpec, error) {
	manualEntries := splitEntries(targetInput)
	cidrEntries := splitEntries(cidrInput)

	if len(manualEntries) == 0 && len(cidrEntries) == 0 {
		return nil, ValidationError{Message: "请至少输入一个目标地址或网段"}
	}

	targets := make([]targetSpec, 0, len(manualEntries))
	indexByAddress := make(map[string]int)

	addTarget := func(address, source string) error {
		if existingIndex, ok := indexByAddress[address]; ok {
			targets[existingIndex].Source = mergeSource(targets[existingIndex].Source, source)
			return nil
		}

		if len(targets) >= limit {
			return ValidationError{Message: fmt.Sprintf("目标数量不能超过 %d 个", limit)}
		}

		indexByAddress[address] = len(targets)
		targets = append(targets, targetSpec{
			Address: address,
			Source:  source,
			Index:   len(targets),
		})
		return nil
	}

	for _, entry := range manualEntries {
		expanded, err := expandManualEntry(entry, limit)
		if err != nil {
			return nil, err
		}
		for _, address := range expanded {
			if err := addTarget(address, SourceManual); err != nil {
				return nil, err
			}
		}
	}

	for _, entry := range cidrEntries {
		expanded, err := expandCIDR(entry, limit)
		if err != nil {
			return nil, err
		}
		for _, address := range expanded {
			if err := addTarget(address, SourceCIDR); err != nil {
				return nil, err
			}
		}
	}

	if len(targets) == 0 {
		return nil, ValidationError{Message: "没有可检测的目标"}
	}

	return targets, nil
}

func splitEntries(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	rawEntries := splitPattern.Split(strings.TrimSpace(input), -1)
	entries := make([]string, 0, len(rawEntries))
	for _, entry := range rawEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func isValidTarget(value string) bool {
	if ip := net.ParseIP(value); ip != nil {
		return true
	}
	return hostPattern.MatchString(value) && !strings.Contains(value, "..")
}

func expandManualEntry(value string, limit int) ([]string, error) {
	if match := fullIPv4RangePattern.FindStringSubmatch(value); len(match) == 3 {
		return expandIPRange(match[1], match[2], value, limit)
	}

	if match := shortIPv4RangePattern.FindStringSubmatch(value); len(match) == 3 {
		start, err := netip.ParseAddr(match[1])
		if err != nil || !start.Is4() {
			return nil, ValidationError{Message: fmt.Sprintf("无效目标范围: %s", value)}
		}

		endOctet, err := strconv.Atoi(match[2])
		if err != nil || endOctet < 0 || endOctet > 255 {
			return nil, ValidationError{Message: fmt.Sprintf("无效目标范围: %s", value)}
		}

		endBytes := start.As4()
		endBytes[3] = byte(endOctet)
		end := netip.AddrFrom4(endBytes)
		return expandIPRange(start.String(), end.String(), value, limit)
	}

	if match := fullIPv6RangePattern.FindStringSubmatch(value); len(match) == 3 {
		return expandIPRange(match[1], match[2], value, limit)
	}

	if !isValidTarget(value) {
		return nil, ValidationError{Message: fmt.Sprintf("无效目标: %s", value)}
	}

	return []string{value}, nil
}

func expandCIDR(value string, limit int) ([]string, error) {
	prefix, err := netip.ParsePrefix(value)
	if err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("无效网段: %s", value)}
	}

	prefix = prefix.Masked()
	if prefix.Addr().Is4() {
		return expandIPv4CIDR(prefix, limit)
	}
	return expandIPv6CIDR(prefix, limit)
}

func expandIPv4CIDR(prefix netip.Prefix, limit int) ([]string, error) {
	hostBits := 32 - prefix.Bits()
	total := uint64(1) << hostBits
	usable := total
	if total > 2 {
		usable = total - 2
	}

	if usable > uint64(limit) {
		return nil, ValidationError{Message: fmt.Sprintf("网段 %s 主机数过多，请控制在 %d 个以内", prefix.String(), limit)}
	}

	base := ipv4ToUint32(prefix.Addr())
	start := base
	end := base + uint32(total) - 1
	if total > 2 {
		start++
		end--
	}

	values := make([]string, 0, int(usable))
	for current := start; ; current++ {
		values = append(values, uint32ToIPv4(current))
		if current == end {
			break
		}
	}

	return values, nil
}

func expandIPv6CIDR(prefix netip.Prefix, limit int) ([]string, error) {
	values := make([]string, 0, minInt(limit, 256))
	for addr := prefix.Addr(); prefix.Contains(addr); {
		if len(values) >= limit {
			return nil, ValidationError{Message: fmt.Sprintf("网段 %s 主机数过多，请控制在 %d 个以内", prefix.String(), limit)}
		}

		values = append(values, addr.String())
		next := addr.Next()
		if !next.IsValid() {
			break
		}
		addr = next
	}

	return values, nil
}

func expandIPRange(startText, endText, original string, limit int) ([]string, error) {
	start, err := netip.ParseAddr(startText)
	if err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("无效目标范围: %s", original)}
	}

	end, err := netip.ParseAddr(endText)
	if err != nil {
		return nil, ValidationError{Message: fmt.Sprintf("无效目标范围: %s", original)}
	}

	if start.Is4() != end.Is4() {
		return nil, ValidationError{Message: fmt.Sprintf("目标范围起始和结束地址类型不一致: %s", original)}
	}

	if start.Compare(end) > 0 {
		return nil, ValidationError{Message: fmt.Sprintf("目标范围起始值不能大于结束值: %s", original)}
	}

	values := make([]string, 0, minInt(limit, 256))
	for current := start; ; {
		if len(values) >= limit {
			return nil, ValidationError{Message: fmt.Sprintf("目标范围 %s 数量过多，请控制在 %d 个以内", original, limit)}
		}

		values = append(values, current.String())
		if current.Compare(end) == 0 {
			break
		}

		next := current.Next()
		if !next.IsValid() {
			return nil, ValidationError{Message: fmt.Sprintf("无效目标范围: %s", original)}
		}
		current = next
	}

	return values, nil
}

func ipv4ToUint32(addr netip.Addr) uint32 {
	bytes := addr.As4()
	return binary.BigEndian.Uint32(bytes[:])
}

func uint32ToIPv4(value uint32) string {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], value)
	return netip.AddrFrom4(bytes).String()
}

func mergeSource(existing, next string) string {
	if existing == next || strings.Contains(existing, next) {
		return existing
	}

	if existing == SourceCIDR && next == SourceManual {
		return SourceManual + "," + SourceCIDR
	}

	if existing == SourceManual && next == SourceCIDR {
		return SourceManual + "," + SourceCIDR
	}

	return existing + "," + next
}
