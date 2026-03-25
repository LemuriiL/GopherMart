package luhn

import "testing"

func TestIsDigits(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "digits", value: "123456", want: true},
		{name: "empty", value: "", want: false},
		{name: "letters", value: "12a34", want: false},
		{name: "spaces", value: "12 34", want: false},
		{name: "symbols", value: "12-34", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDigits(tt.value)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{name: "valid 79927398713", number: "79927398713", want: true},
		{name: "valid 12345678903", number: "12345678903", want: true},
		{name: "invalid short", number: "12345", want: false},
		{name: "invalid random", number: "11111111111", want: false},
		{name: "valid 4242424242424242", number: "4242424242424242", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValid(tt.number)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
