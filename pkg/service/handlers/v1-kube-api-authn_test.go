package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	kubeapiauth "github.com/rancher/kube-api-auth/pkg"
	"github.com/rancher/kube-api-auth/pkg/api/v1/types"
)

func TestV1parseBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		token      string
		wantKey    string
		wantSecret string
		wantErr    bool
	}{
		{
			name:       "legacy token",
			token:      "tokenName:secretValue",
			wantKey:    "tokenName",
			wantSecret: "secretValue",
		},
		{
			name:       "ext token",
			token:      "ext/token-abc123:secretValue",
			wantKey:    "token-abc123",
			wantSecret: "secretValue",
		},
		{
			name:       "ext token with colon in secret",
			token:      "ext/token-abc123:secret:with:colons",
			wantKey:    "token-abc123",
			wantSecret: "secret:with:colons",
		},
		{
			name:       "legacy token with colon in secret",
			token:      "tokenName:secret:with:colons",
			wantKey:    "tokenName",
			wantSecret: "secret:with:colons",
		},
		{
			name:    "missing colon",
			token:   "tokenNameOnly",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body := types.V1AuthnRequest{
				APIVersion: kubeapiauth.DefaultK8sAPIVersion,
				Kind:       kubeapiauth.DefaultAuthnKind,
				Spec:       types.V1AuthnRequestSpec{Token: tt.token},
			}
			data, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			r := httptest.NewRequest("POST", "/v1/authenticate", bytes.NewReader(data))
			accessKey, secretKey, err := v1parseBody(r)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got accessKey=%q secretKey=%q", accessKey, secretKey)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if accessKey != tt.wantKey {
				t.Errorf("accessKey = %q, want %q", accessKey, tt.wantKey)
			}
			if secretKey != tt.wantSecret {
				t.Errorf("secretKey = %q, want %q", secretKey, tt.wantSecret)
			}
		})
	}
}
