package config

import (
	"os"
	"path/filepath"

	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterConfigFile(t *testing.T) {
	GatewayConfigDirName = "../../test/configs/gateway/"

	filePath := filepath.Join(GatewayConfigDirName, GatewayConfigFileName)
	_, err := os.Stat(filePath)
	require.NoError(t, err)

	gatewayConfig, err := RegisterConfigFile()
	require.NoError(t, err)

	expectedConfig := createDummyGatewayConfig()
	require.Equal(t, expectedConfig, gatewayConfig)
}

func TestUpdateNginxConfig(t *testing.T) {
	cfg := createDummyGatewayConfig()
	err := UpdateNginxConfig(NGINXConfigDirName+NGINXConfigFileName, "", cfg)
	require.NoError(t, err)

	// read the new file
	file, err := os.ReadFile(NGINXConfigDirName + NGINXConfigFileName)
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

func removeFilePath(t *testing.T, filename string) {
	err := os.RemoveAll(filename)
	require.NoError(t, err)
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
