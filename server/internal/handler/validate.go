package handler

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// uciSafe ensures a string is safe for embedding in UCI config values.
// Rejects single quotes, newlines, and shell metacharacters.
var uciUnsafeChars = regexp.MustCompile(`['\n\r\\` + "`" + `$;|&><]`)

func validateUCIValue(field, value string) error {
	if uciUnsafeChars.MatchString(value) {
		return fmt.Errorf("%s contains invalid characters", field)
	}
	return nil
}

// validateIP checks that a string is a valid IPv4 or IPv6 address.
func validateIP(field, value string) error {
	if value == "" {
		return nil
	}
	if net.ParseIP(value) == nil {
		return fmt.Errorf("%s is not a valid IP address", field)
	}
	return nil
}

// validateCIDR checks that a string is a valid CIDR notation.
func validateCIDR(field, value string) error {
	if value == "" {
		return nil
	}
	_, _, err := net.ParseCIDR(value)
	if err != nil {
		return fmt.Errorf("%s is not a valid CIDR", field)
	}
	return nil
}

// validatePort checks that a string is a valid port or port range.
var portRangeRE = regexp.MustCompile(`^(\d+)(-(\d+))?$`)

func validatePort(field, value string) error {
	if value == "" {
		return nil
	}
	// Ports can be comma-separated: "80,443" or ranges: "8000-9000"
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		m := portRangeRE.FindStringSubmatch(part)
		if m == nil {
			return fmt.Errorf("%s contains invalid port: %s", field, part)
		}
		p, _ := strconv.Atoi(m[1])
		if p < 1 || p > 65535 {
			return fmt.Errorf("%s port out of range: %d", field, p)
		}
		if m[3] != "" {
			p2, _ := strconv.Atoi(m[3])
			if p2 < 1 || p2 > 65535 || p2 < p {
				return fmt.Errorf("%s port range invalid: %s", field, part)
			}
		}
	}
	return nil
}

// validateOneOf checks that value is one of the allowed strings (case-insensitive).
func validateOneOf(field, value string, allowed []string) error {
	v := strings.ToUpper(value)
	for _, a := range allowed {
		if v == strings.ToUpper(a) {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", field, strings.Join(allowed, ", "))
}

// validatePassword enforces complexity: min 8 chars, at least one upper, one lower, one digit.
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}
	return nil
}

var firewallTargets = []string{"ACCEPT", "REJECT", "DROP"}
var firewallProtos = []string{"tcp", "udp", "tcp udp", "icmp", "any"}
var mwanProtos = []string{"all", "tcp", "udp", "icmp"}

// validateMAC checks MAC address format.
var macRE = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)

func validateMAC(field, value string) error {
	if value == "" {
		return nil
	}
	if !macRE.MatchString(value) {
		return fmt.Errorf("%s is not a valid MAC address", field)
	}
	return nil
}

// validateName checks that a name only contains safe characters for UCI identifiers.
var nameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateName(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !nameRE.MatchString(value) {
		return fmt.Errorf("%s must contain only letters, digits, hyphens, and underscores", field)
	}
	return nil
}
