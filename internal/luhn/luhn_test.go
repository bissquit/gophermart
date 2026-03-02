package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Valid(t *testing.T) {
	type want struct {
		isValid bool
	}
	tests := []struct {
		name   string
		number string
		want   want
	}{
		{
			name:   "valid number",
			number: "17893729974",
			want: want{
				isValid: true,
			},
		},
		{
			name:   "empty number",
			number: "",
			want: want{
				isValid: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := Valid(tt.number)
			assert.Equal(t, valid, tt.want.isValid)
		})
	}
}
