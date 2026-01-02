package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterConfigFile(t *testing.T) {
	err := os.MkdirAll(GatewayConfigDirName, 0755)
	require.NoError(t, err)

	filePath := filepath.Join(GatewayConfigDirName, GatewayConfigFileName)
	file, err := os.Create(filePath)
	require.NoError(t, err)

	defer removeFilePath(t, GatewayConfigDirName)

	fileBody, err := os.ReadFile("../../test/configs/gateway.yaml")
	require.NoError(t, err)

	err = os.WriteFile(file.Name(), fileBody, 0755)
	require.NoError(t, err)

	err = file.Close()
	require.NoError(t, err)

	_, err = RegisterConfigFile()
	require.NoError(t, err)
}

func TestUpdateNginxConfig(t *testing.T) {
	cfg := createDummyGatewayConfig()
	err := UpdateNginxConfig(NGINXConfigDirName+NGINXConfigFileName, "", &cfg)
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

func createDummyGatewayConfig() GatewayConfig {
	return GatewayConfig{
		Connection: []Connection{
			{
				Host: "localhost",
				Port: 8080,
				Routes: []Route{
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
