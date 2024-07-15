package jwt

import (
	"reflect"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJwt(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Empty Username",
			args:    args{username: ""},
			wantErr: false,
		},
		{
			name:    "Non-empty Username",
			args:    args{username: "testUser"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJwt(tt.args.username)
			t.Logf("Generated jwt: %s", token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJwt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseJwt(t *testing.T) {
	validUsername1 := "xiaoxinzi"
	validToken1, err := GenerateJwt(validUsername1)
	if err != nil {
		t.Fatalf("GenerateJwt() error = %v", err)
	}
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		want    *jwt.RegisteredClaims
		wantErr bool
	}{
		{
			name:    "Valid token",
			args:    args{tokenString: validToken1},
			want:    &jwt.RegisteredClaims{Subject: validUsername1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJwt(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJwt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJwt() = %v, want %v", got, tt.want)
			}
		})
	}
}
