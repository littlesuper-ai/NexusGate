package handler

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "Abcdef1234", false},
		{"minimum valid", "Abcdefg1", false},
		{"too short", "Abc1", true},
		{"no uppercase", "abcdefg1", true},
		{"no lowercase", "ABCDEFG1", true},
		{"no digit", "Abcdefgh", true},
		{"empty", "", true},
		{"only digits", "12345678", true},
		{"only lowercase", "abcdefgh", true},
		{"only uppercase", "ABCDEFGH", true},
		{"with special chars", "Abc!@#$1", false},
		{"unicode with requirements", "Abc12345你好", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateUCIValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"normal string", "my-value", false},
		{"alphanumeric", "abc123", false},
		{"with single quote", "it's", true},
		{"with newline", "line1\nline2", true},
		{"with semicolon", "cmd; rm -rf", true},
		{"with pipe", "a | b", true},
		{"with ampersand", "a && b", true},
		{"with dollar", "$HOME", true},
		{"with backtick", "`whoami`", true},
		{"with backslash", "a\\b", true},
		{"empty string", "", false},
		{"with spaces", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUCIValue("field", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv6", "::1", false},
		{"valid IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"invalid IP", "999.999.999.999", true},
		{"text", "not-an-ip", true},
		{"partial", "192.168", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIP("ip", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"valid CIDR", "192.168.1.0/24", false},
		{"valid /32", "10.0.0.1/32", false},
		{"valid IPv6 CIDR", "2001:db8::/32", false},
		{"missing mask", "192.168.1.0", true},
		{"invalid CIDR", "999.0.0.0/8", true},
		{"text", "not-a-cidr", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCIDR("cidr", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"single port", "80", false},
		{"max port", "65535", false},
		{"port range", "8000-9000", false},
		{"comma separated", "80,443", false},
		{"mixed", "80,8000-9000,443", false},
		{"port zero", "0", true},
		{"port too high", "65536", true},
		{"invalid range", "9000-8000", true},
		{"text", "http", true},
		{"negative", "-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort("port", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateOneOf(t *testing.T) {
	allowed := []string{"ACCEPT", "REJECT", "DROP"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"exact match", "ACCEPT", false},
		{"case insensitive", "accept", false},
		{"mixed case", "Reject", false},
		{"invalid", "FORWARD", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOneOf("target", tt.value, allowed)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMAC(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"valid MAC", "AA:BB:CC:DD:EE:FF", false},
		{"lowercase MAC", "aa:bb:cc:dd:ee:ff", false},
		{"mixed case", "Aa:Bb:Cc:Dd:Ee:Ff", false},
		{"no colons", "AABBCCDDEEFF", true},
		{"short", "AA:BB:CC", true},
		{"invalid chars", "GG:HH:II:JJ:KK:LL", true},
		{"dash separated", "AA-BB-CC-DD-EE-FF", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMAC("mac", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid name", "my-device_01", false},
		{"simple", "router1", false},
		{"empty is invalid", "", true},
		{"with spaces", "my device", true},
		{"with special chars", "dev@home", true},
		{"with dots", "dev.local", true},
		{"underscore only", "___", false},
		{"hyphen only", "---", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName("name", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
