package config

import (
	"fyp-api-gateway/src/utils"
	"os"
	"path/filepath"

	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterConfigFile(t *testing.T) {
	utils.GatewayConfigDirName = "../../test/configs/gateway/"

	filePath := filepath.Join(utils.GatewayConfigDirName, utils.GatewayConfigFileName)
	_, err := os.Stat(filePath)
	require.NoError(t, err)

	utils.NGINXTemplateDirName = "../templates/"
	utils.NGINXDirName = "../../test/configs/nginx/"
	store := NewConfigStore()
	gatewayConfig, err := RegisterConfigFile(store)
	require.NoError(t, err)

	expectedConfig := createDummyGatewayConfig()
	require.Equal(t, expectedConfig, gatewayConfig)
}

func TestUpdateNginxConfig(t *testing.T) {
	cfg := createDummyGatewayConfig()
	store := NewConfigStore()

	err := UpdateNginxConfig(utils.NGINXDirName+utils.NGINXConfigFileName, "", cfg, store)
	require.NoError(t, err)

	// read the new file
	file, err := os.ReadFile(utils.NGINXDirName + utils.NGINXConfigFileName)
	require.NoError(t, err)
	expectedFile, err := os.ReadFile("../../test/configs/nginx/nginx.conf")
	require.NoError(t, err)
	require.Equal(t,
		removeWhitespace(string(expectedFile)),
		removeWhitespace(string(file)),
	)
}

func removeWhitespace(f string) string {
	lines := strings.Split(f, "\n")
	var out []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n")
}

func createDummyGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		Connections: []Connections{
			{
				Host: "localhost",
				Port: 8080,
				Routes: []Routes{
					{
						Path: "/products",
						Upstream: Upstream{
							Name: "product_service",
							Port: 9001,
						},
					},
					{
						Path: "/orders",
						Upstream: Upstream{
							Name: "order_service",
							Port: 9002,
						},
					},
				},
			},
		},
	}
}
