package validator

import "testing"

func TestNormalizePhone_Chinese(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"bare 11-digit china", "13800138000", "+8613800138000", false},
		{"country prefix 86", "8613800138000", "+8613800138000", false},
		{"with plus", "+8613800138000", "+8613800138000", false},
		{"with formatting", "+86-138-0013-8000", "+8613800138000", false},
		{"with spaces", "138 0013 8000", "+8613800138000", false},
		{"empty", "", "", true},
		{"no digits", "---", "", true},
		{"too short", "12345", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePhone(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizePhone(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NormalizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizePhone_International(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"US number with plus", "+1 650 555 0199", "+16505550199"},
		{"UK number", "+44 20 7946 0958", "+442079460958"},
		{"Japan number", "+81 90 1234 5678", "+819012345678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePhone(tt.input)
			if err != nil {
				t.Fatalf("NormalizePhone(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"bare china", "13800138000", true},
		{"e164 china", "+8613800138000", true},
		{"formatted china", "+86-138-0013-8000", true},
		{"US e164", "+16505550199", true},
		{"empty", "", false},
		{"too short", "12345", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePhone(tt.input); got != tt.want {
				t.Fatalf("ValidatePhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
