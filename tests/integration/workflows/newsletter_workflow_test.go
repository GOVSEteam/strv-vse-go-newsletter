package workflows_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteNewsletterWorkflow tests the entire newsletter workflow from creation to publishing
func TestCompleteNewsletterWorkflow(t *testing.T) {
	// Setup test server and dependencies
	testServer := setup.NewTestServer(t)
	defer testServer.Close()

	// Setup authentication
	authConfig := setup.DefaultTestAuthConfig()
	authHelper := setup.NewTestAuthHelper(t, authConfig)
	defer authHelper.Cleanup()

	// Create test editor in database
	editorID := authHelper.CreateTestEditor(t, testServer.DB)
	t.Logf("Created test editor with ID: %s", editorID)

	baseURL := testServer.URL()

	// Step 1: Create a newsletter
	t.Run("Create Newsletter", func(t *testing.T) {
		newsletterData := map[string]interface{}{
			"name":        "Integration Test Newsletter",
			"description": "A newsletter created during integration testing",
		}

		body, err := json.Marshal(newsletterData)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/api/newsletters", bytes.NewBuffer(body))
		require.NoError(t, err)
		authHelper.AddAuthHeaders(req)

		resp, err := testServer.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var newsletter map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&newsletter)
		require.NoError(t, err)

		assert.Equal(t, "Integration Test Newsletter", newsletter["name"])
		assert.Equal(t, "A newsletter created during integration testing", newsletter["description"])
		assert.NotEmpty(t, newsletter["id"])

		// Store newsletter ID for next steps
		newsletterID := newsletter["id"].(string)
		t.Logf("Created newsletter with ID: %s", newsletterID)

		// Step 2: Create a post for the newsletter
		t.Run("Create Post", func(t *testing.T) {
			postData := map[string]interface{}{
				"title":   "Welcome to Our Newsletter",
				"content": "This is the first post in our integration test newsletter. It contains important information for our subscribers.",
			}

			body, err := json.Marshal(postData)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", baseURL+"/api/newsletters/"+newsletterID+"/posts", bytes.NewBuffer(body))
			require.NoError(t, err)
			authHelper.AddAuthHeaders(req)

			resp, err := testServer.Client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var post map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&post)
			require.NoError(t, err)

			assert.Equal(t, "Welcome to Our Newsletter", post["title"])
			assert.Equal(t, "This is the first post in our integration test newsletter. It contains important information for our subscribers.", post["content"])
			assert.Nil(t, post["published_at"])
			assert.NotEmpty(t, post["id"])

			postID := post["id"].(string)
			t.Logf("Created post with ID: %s", postID)

			// Step 3: Update the post
			t.Run("Update Post", func(t *testing.T) {
				updateData := map[string]interface{}{
					"title":   "Welcome to Our Newsletter - Updated",
					"content": "This is the updated content of our first post. We've made some improvements based on feedback.",
				}

				body, err := json.Marshal(updateData)
				require.NoError(t, err)

				req, err := http.NewRequest("PUT", baseURL+"/api/posts/"+postID, bytes.NewBuffer(body))
				require.NoError(t, err)
				authHelper.AddAuthHeaders(req)

				resp, err := testServer.Client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var updatedPost map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&updatedPost)
				require.NoError(t, err)

				assert.Equal(t, "Welcome to Our Newsletter - Updated", updatedPost["title"])
				assert.Equal(t, "This is the updated content of our first post. We've made some improvements based on feedback.", updatedPost["content"])

				// Step 4: Get the post to verify update
				t.Run("Get Updated Post", func(t *testing.T) {
					req, err := http.NewRequest("GET", baseURL+"/api/posts/"+postID, nil)
					require.NoError(t, err)
					authHelper.AddAuthHeaders(req)

					resp, err := testServer.Client.Do(req)
					require.NoError(t, err)
					defer resp.Body.Close()

					assert.Equal(t, http.StatusOK, resp.StatusCode)

					var retrievedPost map[string]interface{}
					err = json.NewDecoder(resp.Body).Decode(&retrievedPost)
					require.NoError(t, err)

					assert.Equal(t, "Welcome to Our Newsletter - Updated", retrievedPost["title"])
					assert.Equal(t, "This is the updated content of our first post. We've made some improvements based on feedback.", retrievedPost["content"])

					// Step 5: List posts for the newsletter
					t.Run("List Newsletter Posts", func(t *testing.T) {
						req, err := http.NewRequest("GET", baseURL+"/api/newsletters/"+newsletterID+"/posts", nil)
						require.NoError(t, err)
						authHelper.AddAuthHeaders(req)

						resp, err := testServer.Client.Do(req)
						require.NoError(t, err)
						defer resp.Body.Close()

						assert.Equal(t, http.StatusOK, resp.StatusCode)

						var postsResponse map[string]interface{}
						err = json.NewDecoder(resp.Body).Decode(&postsResponse)
						require.NoError(t, err)

						assert.Contains(t, postsResponse, "data")
						assert.Contains(t, postsResponse, "total")
						assert.Contains(t, postsResponse, "limit")
						assert.Contains(t, postsResponse, "offset")

						posts := postsResponse["data"].([]interface{})
						assert.Len(t, posts, 1)

						firstPost := posts[0].(map[string]interface{})
						assert.Equal(t, "Welcome to Our Newsletter - Updated", firstPost["title"])

						// Step 6: Publish the post (if publishing service is available)
						t.Run("Publish Post", func(t *testing.T) {
							req, err := http.NewRequest("POST", baseURL+"/api/posts/"+postID+"/publish", nil)
							require.NoError(t, err)
							authHelper.AddAuthHeaders(req)

							resp, err := testServer.Client.Do(req)
							require.NoError(t, err)
							defer resp.Body.Close()

							// Publishing might fail if no subscribers exist, but the endpoint should respond
							assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)

							if resp.StatusCode == http.StatusOK {
								var publishResponse map[string]interface{}
								err = json.NewDecoder(resp.Body).Decode(&publishResponse)
								require.NoError(t, err)
								t.Logf("Post published successfully: %v", publishResponse)
							} else {
								var errorResponse map[string]interface{}
								err = json.NewDecoder(resp.Body).Decode(&errorResponse)
								require.NoError(t, err)
								t.Logf("Publishing failed as expected (no subscribers): %v", errorResponse)
							}
						})
					})
				})
			})
		})
	})
}

// TestNewsletterCRUDWorkflow tests the complete CRUD operations for newsletters
func TestNewsletterCRUDWorkflow(t *testing.T) {
	// Setup test server and dependencies
	testServer := setup.NewTestServer(t)
	defer testServer.Close()

	// Setup authentication
	authConfig := setup.DefaultTestAuthConfig()
	authHelper := setup.NewTestAuthHelper(t, authConfig)
	defer authHelper.Cleanup()

	// Create test editor in database
	editorID := authHelper.CreateTestEditor(t, testServer.DB)
	t.Logf("Created test editor with ID: %s", editorID)

	baseURL := testServer.URL()
	var newsletterID string

	// Create newsletter
	t.Run("Create Newsletter", func(t *testing.T) {
		newsletterData := map[string]interface{}{
			"name":        "CRUD Test Newsletter",
			"description": "Testing CRUD operations",
		}

		body, err := json.Marshal(newsletterData)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/api/newsletters", bytes.NewBuffer(body))
		require.NoError(t, err)
		authHelper.AddAuthHeaders(req)

		resp, err := testServer.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var newsletter map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&newsletter)
		require.NoError(t, err)

		newsletterID = newsletter["id"].(string)
		assert.NotEmpty(t, newsletterID)
	})

	// Update newsletter
	t.Run("Update Newsletter", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name":        "CRUD Test Newsletter - Updated",
			"description": "Testing CRUD operations - Updated description",
		}

		body, err := json.Marshal(updateData)
		require.NoError(t, err)

		req, err := http.NewRequest("PATCH", baseURL+"/api/newsletters/"+newsletterID, bytes.NewBuffer(body))
		require.NoError(t, err)
		authHelper.AddAuthHeaders(req)

		resp, err := testServer.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedNewsletter map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&updatedNewsletter)
		require.NoError(t, err)

		assert.Equal(t, "CRUD Test Newsletter - Updated", updatedNewsletter["name"])
		assert.Equal(t, "Testing CRUD operations - Updated description", updatedNewsletter["description"])
	})

	// List newsletters
	t.Run("List Newsletters", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURL+"/api/newsletters", nil)
		require.NoError(t, err)
		authHelper.AddAuthHeaders(req)

		resp, err := testServer.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var newslettersResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&newslettersResponse)
		require.NoError(t, err)

		assert.Contains(t, newslettersResponse, "data")
		newsletters := newslettersResponse["data"].([]interface{})
		assert.GreaterOrEqual(t, len(newsletters), 1)

		// Find our newsletter
		found := false
		for _, nl := range newsletters {
			newsletter := nl.(map[string]interface{})
			if newsletter["id"].(string) == newsletterID {
				found = true
				assert.Equal(t, "CRUD Test Newsletter - Updated", newsletter["name"])
				break
			}
		}
		assert.True(t, found, "Created newsletter should be in the list")
	})

	// Delete newsletter
	t.Run("Delete Newsletter", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", baseURL+"/api/newsletters/"+newsletterID, nil)
		require.NoError(t, err)
		authHelper.AddAuthHeaders(req)

		resp, err := testServer.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode) // DELETE typically returns 204

		// Verify deletion by trying to get the newsletter list (should not contain deleted newsletter)
		req, err = http.NewRequest("GET", baseURL+"/api/newsletters", nil)
		require.NoError(t, err)
		authHelper.AddAuthHeaders(req)

		resp, err = testServer.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var newslettersResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&newslettersResponse)
		require.NoError(t, err)

		if newslettersResponse["data"] != nil {
			newsletters := newslettersResponse["data"].([]interface{})
			for _, nl := range newsletters {
				newsletter := nl.(map[string]interface{})
				assert.NotEqual(t, newsletterID, newsletter["id"].(string), "Deleted newsletter should not be in the list")
			}
		}
	})
} 