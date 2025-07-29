package dto

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestRegisterRequestModel(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		isError bool
	}{
		{
			name: "successfully case",
			req: RegisterRequest{
				Login:    "hello",
				Password: "superpass",
			},

			isError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.req.ToModel()
			if !tt.isError && err != nil {
				t.Errorf("expected no error")
			}
			if result.Login != tt.req.Login ||
				!errors.Is(bcrypt.CompareHashAndPassword([]byte(result.Password),
					[]byte(tt.req.Password)), nil) {
				t.Errorf("bad result")
			}
		})
	}
}
