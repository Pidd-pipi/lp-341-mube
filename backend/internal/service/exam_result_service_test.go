package service

import "testing"

func TestIsAbnormal(t *testing.T) {
	tests := []struct {
		name     string
		refRange string
		value    string
		want     bool
	}{
		{name: "in range", refRange: "3.5-9.5", value: "5.2", want: false},
		{name: "low", refRange: "3.5-9.5", value: "2.1", want: true},
		{name: "high", refRange: "3.5-9.5", value: "12.0", want: true},
		{name: "greater-than", refRange: ">10", value: "9", want: true},
		{name: "greater-than ok", refRange: ">10", value: "11", want: false},
		{name: "less-than", refRange: "<5", value: "6", want: true},
		{name: "normal text", refRange: "正常", value: "正常", want: false},
		{name: "empty", refRange: "", value: "5", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAbnormal(tt.refRange, tt.value); got != tt.want {
				t.Fatalf("IsAbnormal(%q, %q) = %v, want %v", tt.refRange, tt.value, got, tt.want)
			}
		})
	}
}

func TestGuessAbnormalLevel(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "72", want: "mild"},
		{value: "500", want: "moderate"},
		{value: "5000", want: "severe"},
		{value: "abc", want: "mild"},
	}
	for _, tt := range tests {
		if got := GuessAbnormalLevel(tt.value); got != tt.want {
			t.Fatalf("GuessAbnormalLevel(%q) = %s, want %s", tt.value, got, tt.want)
		}
	}
}
