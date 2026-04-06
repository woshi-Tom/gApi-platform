package model

import (
	"testing"
	"time"
)

func TestGenerateCode(t *testing.T) {
	prefix := "TEST"
	code := GenerateCode(prefix)

	if len(code) < len(prefix) {
		t.Errorf("GenerateCode() returned code shorter than prefix: %s", code)
	}

	if code[:len(prefix)] != prefix {
		t.Errorf("GenerateCode() prefix mismatch: got %s, expected prefix %s", code[:len(prefix)], prefix)
	}

	// Check length (prefix + timestamp(10) + suffix(4))
	expectedMinLength := len(prefix) + 14
	if len(code) < expectedMinLength {
		t.Errorf("GenerateCode() returned code too short: %s (length %d, expected >= %d)", code, len(code), expectedMinLength)
	}
}

func TestRedemptionCode_IsValid(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	pastTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		code     RedemptionCode
		expected bool
	}{
		{
			name: "active code with available uses",
			code: RedemptionCode{
				Status:    RedemptionStatusActive,
				UsedCount: 0,
				MaxUses:   1,
			},
			expected: true,
		},
		{
			name: "disabled code",
			code: RedemptionCode{
				Status:    RedemptionStatusDisabled,
				UsedCount: 0,
				MaxUses:   1,
			},
			expected: false,
		},
		{
			name: "used up code",
			code: RedemptionCode{
				Status:    RedemptionStatusActive,
				UsedCount: 1,
				MaxUses:   1,
			},
			expected: false,
		},
		{
			name: "code with multiple uses available",
			code: RedemptionCode{
				Status:    RedemptionStatusActive,
				UsedCount: 2,
				MaxUses:   10,
			},
			expected: true,
		},
		{
			name: "code not yet valid (valid_from in future)",
			code: RedemptionCode{
				Status:    RedemptionStatusActive,
				UsedCount: 0,
				MaxUses:   1,
				ValidFrom: &futureTime,
			},
			expected: false,
		},
		{
			name: "code expired (valid_until in past)",
			code: RedemptionCode{
				Status:     RedemptionStatusActive,
				UsedCount:  0,
				MaxUses:    1,
				ValidUntil: &pastTime,
			},
			expected: false,
		},
		{
			name: "code with valid date range",
			code: RedemptionCode{
				Status:     RedemptionStatusActive,
				UsedCount:  0,
				MaxUses:    1,
				ValidFrom:  &pastTime,
				ValidUntil: &futureTime,
			},
			expected: true,
		},
		{
			name: "used status code",
			code: RedemptionCode{
				Status:    RedemptionStatusUsed,
				UsedCount: 0,
				MaxUses:   1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.code.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateCode_Uniqueness(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 10; i++ {
		code := GenerateCode("UNIQ")
		if codes[code] {
			t.Errorf("GenerateCode() produced duplicate code: %s", code)
		}
		codes[code] = true
		time.Sleep(time.Millisecond)
	}
}

func TestRedemptionCodeTypes(t *testing.T) {
	// Test that constants are defined correctly
	tests := []struct {
		codeType string
		expected string
	}{
		{RedemptionCodeTypeRecharge, "recharge"},
		{RedemptionCodeTypeVIP, "vip"},
		{RedemptionCodeTypeQuota, "quota"},
	}

	for _, tt := range tests {
		t.Run(tt.codeType, func(t *testing.T) {
			if tt.codeType != tt.expected {
				t.Errorf("RedemptionCodeType = %s, expected %s", tt.codeType, tt.expected)
			}
		})
	}
}

func TestRedemptionStatus(t *testing.T) {
	// Test status constants
	tests := []struct {
		status   string
		expected string
	}{
		{RedemptionStatusActive, "active"},
		{RedemptionStatusUsed, "used"},
		{RedemptionStatusDisabled, "disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("RedemptionStatus = %s, expected %s", tt.status, tt.expected)
			}
		})
	}
}
