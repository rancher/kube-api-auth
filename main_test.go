package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kubeapiauth "github.com/rancher/kube-api-auth/pkg"
)

func TestCheckArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "no arguments",
		},
		{
			name: "legacy command alone",
			args: []string{"serve"},
		},
		{
			name:    "flag that used to exist",
			args:    []string{"--listen", "0.0.0.0:6440"},
			wantErr: "unrecognized arguments [--listen 0.0.0.0:6440], kube-api-auth is configured through CATTLE_DEBUG (or RANCHER_DEBUG), KUBECONFIG, CATTLE_NAMESPACE and CATTLE_LISTEN",
		},
		{
			name:    "flags after the legacy command",
			args:    []string{"serve", "--namespace", "other"},
			wantErr: "unrecognized arguments [--namespace other], kube-api-auth is configured through CATTLE_DEBUG (or RANCHER_DEBUG), KUBECONFIG, CATTLE_NAMESPACE and CATTLE_LISTEN",
		},
		{
			name:    "version",
			args:    []string{"--version"},
			wantErr: "unrecognized arguments [--version], kube-api-auth is configured through CATTLE_DEBUG (or RANCHER_DEBUG), KUBECONFIG, CATTLE_NAMESPACE and CATTLE_LISTEN",
		},
		{
			name:    "unknown command",
			args:    []string{"bogus"},
			wantErr: "unrecognized arguments [bogus], kube-api-auth is configured through CATTLE_DEBUG (or RANCHER_DEBUG), KUBECONFIG, CATTLE_NAMESPACE and CATTLE_LISTEN",
		},
		{
			name:    "legacy command in a later position",
			args:    []string{"--debug", "serve"},
			wantErr: "unrecognized arguments [--debug serve], kube-api-auth is configured through CATTLE_DEBUG (or RANCHER_DEBUG), KUBECONFIG, CATTLE_NAMESPACE and CATTLE_LISTEN",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := checkArgs(test.args)

			if test.wantErr != "" {
				assert.EqualError(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Parallel()

	defaults := config{
		Listen:    kubeapiauth.DefaultListenHostPort,
		Namespace: kubeapiauth.DefaultNamespace,
	}

	tests := []struct {
		name string
		env  map[string]string
		want config
	}{
		{
			name: "empty environment falls back to defaults",
			want: defaults,
		},
		{
			name: "kubeconfig",
			env:  map[string]string{"KUBECONFIG": "/home/user/.kube/config"},
			want: config{
				Listen:     kubeapiauth.DefaultListenHostPort,
				Namespace:  kubeapiauth.DefaultNamespace,
				Kubeconfig: "/home/user/.kube/config",
			},
		},
		{
			name: "listen",
			env:  map[string]string{"CATTLE_LISTEN": "127.0.0.1:9000"},
			want: config{Listen: "127.0.0.1:9000", Namespace: kubeapiauth.DefaultNamespace},
		},
		{
			name: "namespace",
			env:  map[string]string{"CATTLE_NAMESPACE": "other"},
			want: config{Listen: kubeapiauth.DefaultListenHostPort, Namespace: "other"},
		},
		{
			name: "empty values keep the defaults",
			env:  map[string]string{"CATTLE_LISTEN": "", "CATTLE_NAMESPACE": ""},
			want: defaults,
		},
		{
			name: "debug",
			env:  map[string]string{"CATTLE_DEBUG": "true"},
			want: config{Listen: kubeapiauth.DefaultListenHostPort, Namespace: kubeapiauth.DefaultNamespace, Debug: true},
		},
		{
			name: "debug through the rancher variable",
			env:  map[string]string{"RANCHER_DEBUG": "1"},
			want: config{Listen: kubeapiauth.DefaultListenHostPort, Namespace: kubeapiauth.DefaultNamespace, Debug: true},
		},
		{
			name: "debug set to false",
			env:  map[string]string{"CATTLE_DEBUG": "false"},
			want: defaults,
		},
		{
			name: "debug set to something unparsable",
			env:  map[string]string{"CATTLE_DEBUG": "yes please"},
			want: defaults,
		},
		{
			name: "every variable at once",
			env: map[string]string{
				"KUBECONFIG":       "/tmp/kubeconfig",
				"CATTLE_LISTEN":    ":9000",
				"CATTLE_NAMESPACE": "other",
				"CATTLE_DEBUG":     "true",
			},
			want: config{
				Listen:     ":9000",
				Namespace:  "other",
				Kubeconfig: "/tmp/kubeconfig",
				Debug:      true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := configFromEnv(func(key string) string { return test.env[key] })

			assert.Equal(t, test.want, got)
		})
	}
}
