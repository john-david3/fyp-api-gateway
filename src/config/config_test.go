package config

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func tmpDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "config_test_*")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})

	return dir
}

func createTestConfigWithSingleRoute() *GatewayConfig {
	testConf := &GatewayConfig{
		Connections: Connection{
			Routes: []Routes{
				{
					Path:     "products",
					Url:      "http://services:9001",
					Auth:     false,
					ZoneName: "",
					RateLimit: RateLimit{
						Zone: 10,
						Rate: 5,
					},
				},
			},
		},
	}

	return testConf
}

func createTestConfigWithMultipleRoutes() *GatewayConfig {
	testConf := &GatewayConfig{
		Connections: Connection{
			Routes: []Routes{
				{
					Path:     "products",
					Url:      "http://services:9001",
					Auth:     false,
					ZoneName: "",
					RateLimit: RateLimit{
						Zone: 10,
						Rate: 5,
					},
				},
				{
					Path:     "orders",
					Url:      "http://services:9002",
					Auth:     false,
					ZoneName: "",
					RateLimit: RateLimit{
						Zone: 10,
						Rate: 5,
					},
				},
			},
		},
	}

	return testConf
}

type configTest struct {
	name     string
	input    string
	expected *GatewayConfig
	willFail bool
}

func TestLoadAndValidateGatewayConf(t *testing.T) {
	emptyConf := &GatewayConfig{}
	singleRouteConf := createTestConfigWithSingleRoute()
	multiRouteConf := createTestConfigWithMultipleRoutes()

	tests := []configTest{
		{
			name:     "Empty string as input",
			input:    "",
			expected: emptyConf,
			willFail: false,
		},
		{
			name:     "Only comments as input",
			input:    "# this is a comment\n# another comment\n   \n",
			expected: emptyConf,
			willFail: false,
		},
		{
			name:     "Valid Input with single route",
			input:    "connections:\n  routes:\n    - path: products\n      url: http://services:9001\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: false",
			expected: singleRouteConf,
			willFail: false,
		},
		{
			name:     "Valid Input with multiple routes",
			input:    "connections:\n  routes:\n    - path: products\n      url: http://services:9001\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: false\n\n    - path: orders\n      url: http://services:9002\n      rate-limit:\n        zone: 10\n        rate: 5\n      auth: false",
			expected: multiRouteConf,
			willFail: false,
		},
		{
			name:     "Invalid Input",
			input:    "connections:\n  routes:\n  - path: [invalid yaml",
			expected: emptyConf,
			willFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadAndValidateGatewayConf(tt.input)

			if tt.willFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, cfg)
			}

		})
	}
}

type (
	templateDataTest struct {
		name     string
		input    expectedTemplateData
		expected TemplateData
	}

	expectedTemplateData struct {
		username string
		gw       *GatewayConfig
	}
)

func TestBuildTemplateData(t *testing.T) {
	emptyConf := &GatewayConfig{}
	singleRouteConf := createTestConfigWithSingleRoute()

	tests := []templateDataTest{
		{
			name: "Gateway config is empty",
			input: expectedTemplateData{
				username: "janedoe",
				gw:       emptyConf,
			},
			expected: TemplateData{
				Username:    "janedoe",
				Connections: Connection{},
			},
		},
		{
			name: "Gateway config is not empty",
			input: expectedTemplateData{
				username: "janedoe",
				gw:       singleRouteConf,
			},
			expected: TemplateData{
				Username: "janedoe",
				Connections: Connection{
					Routes: []Routes{
						{
							Path:     "products",
							Url:      "http://services:9001",
							Auth:     false,
							ZoneName: "products",
							RateLimit: RateLimit{
								Zone: 10,
								Rate: 5,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildTemplateData(tt.input.username, tt.input.gw)
			require.Equal(t, tt.expected, data)
		})
	}
}

type NewConfigTest struct {
	name     string
	req      *http.Request
	expected int
}

func TestLoadNewConfig(t *testing.T) {
	emptyBody, _ := json.Marshal(ConfRequest{Content: ""})
	payload := ConfRequest{Content: "connections:\n  routes:\n  - path: [broken yaml"}
	b, _ := json.Marshal(payload)

	tests := []NewConfigTest{
		{
			name:     "Method not Allowed",
			req:      httptest.NewRequest(http.MethodGet, "/config", nil),
			expected: http.StatusMethodNotAllowed,
		},
		{
			name:     "Missing Session Cookie",
			req:      httptest.NewRequest(http.MethodPost, "/config", bytes.NewReader(emptyBody)),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "Invalid JSON body",
			req:      httptest.NewRequest(http.MethodPost, "/config", strings.NewReader("{bad json")),
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid content",
			req:      httptest.NewRequest(http.MethodPost, "/config", bytes.NewReader(b)),
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Invalid JSON body" || tt.name == "Invalid content" {
				tt.req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}

			rr := httptest.NewRecorder()
			LoadNewConfig(rr, tt.req)
			require.Equal(t, tt.expected, rr.Code)
		})
	}
}

type (
	AtomicWriteTest struct {
		name  string
		input AtomicTestData
	}

	AtomicTestData struct {
		dir     string
		file    string
		content []byte
	}
)

func TestAtomicWrites(t *testing.T) {
	dir := tmpDir(t)

	tests := []AtomicWriteTest{
		{
			name: "Valid data supplied",
			input: AtomicTestData{
				dir:     dir,
				file:    "nginx.conf",
				content: []byte("server { listen 80; }"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := atomicWrites(tt.input.dir, tt.input.file, tt.input.content)
			require.NoError(t, err)

			got, err := os.ReadFile(filepath.Join(dir, "nginx.conf"))
			require.NoError(t, err)
			require.Equal(t, tt.input.content, got)

			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "nginx-") && strings.HasSuffix(e.Name(), ".conf") {
					require.Fail(t, "temp file left behind", e.Name())
				}
			}
		})
	}
}
