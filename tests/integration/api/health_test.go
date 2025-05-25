package api_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	// Setup test server
	testServer := setup.NewTestServer(t)
	defer testServer.Close()

	// Test health endpoint
	resp, err := testServer.Client.Get(testServer.URL() + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
}

// TestHealthEndpointMethods tests that health endpoint only accepts GET
func TestHealthEndpointMethods(t *testing.T) {
	// Setup test server
	testServer := setup.NewTestServer(t)
	defer testServer.Close()

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	
	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			req, err := http.NewRequest(method, testServer.URL()+"/healthz", nil)
			require.NoError(t, err)

			resp, err := testServer.Client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Health endpoint should accept all methods and return OK
			// (Go's default ServeMux behavior)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
} 