package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/config/model"
)

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		setEnvVars func()
		name       string
		wantErrMsg string
		wantErr    bool
	}{
		{
			name: "test load env should succeed when config parsed successfully",
			setEnvVars: func() {
				config.ResetLoadEnv()
			},
		},
		{
			name: "test load env should fail when config is not parsed successfully",
			setEnvVars: func() {
				config.ResetLoadEnv()
				t.Setenv("ENABLE_TLS", "notabool")
			},
			wantErr: true,
			wantErrMsg: "load env config failed: env: parse error on field \"EnableTLS\" of type \"bool\": " +
				"strconv.ParseBool: parsing \"notabool\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnvVars != nil {
				tt.setEnvVars()
			}

			err := config.LoadEnv(model.BuildInfo{
				Commit:    "someCommit",
				BuildTime: "someBuildTime",
				Version:   "someVersion",
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, "someCommit", config.Env.Commit)
				require.Equal(t, "8000", config.Env.Port)
				require.Equal(t, "go-web-skeleton", config.Env.ServiceName)
			}
		})
	}
}

// A template that shipped a default administrator password would put the same
// one in every project generated from it, and in every deployment that forgot to
// change it. There must be no default.
func TestLoadEnv_SeedAdminCredentialsHaveNoDefault(t *testing.T) {
	config.ResetLoadEnv()

	require.NoError(t, config.LoadEnv(model.BuildInfo{}))

	require.Empty(t, config.Env.SeedAdminUsername)
	require.Empty(t, config.Env.SeedAdminPassword)
}
