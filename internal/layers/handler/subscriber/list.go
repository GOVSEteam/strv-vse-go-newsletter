package subscriber

import (
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/utils"
)

// GetSubscribersHandler handles requests for an editor to list active subscribers of their newsletter.
// GET /api/newsletters/{newsletterID}/subscribers
// Protected endpoint: Requires editor authentication.
func GetSubscribersHandler(
	subscriberService service.SubscriberServiceInterface,
	newsletterRepo repository.NewsletterRepository, // To verify newsletter ownership
	editorRepo repository.EditorRepository,       // To get editor ID from Firebase UID
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Extract Newsletter ID from path
		newsletterID, err := utils.GetIDFromPath(r.URL.Path, "/api/newsletters/", "/subscribers")
		if err != nil {
			handler.JSONError(w, "Invalid newsletter ID in path: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 2. Authenticate Editor (Verify JWT)
		firebaseUID, err := auth.VerifyFirebaseJWT(r) // Pass the request directly
		if err != nil {
			handler.JSONError(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Get internal editor ID from Firebase UID
		editor, err := editorRepo.GetEditorByFirebaseUID(firebaseUID) // Corrected: removed ctx
		if err != nil {
			handler.JSONError(w, "Failed to retrieve editor details: "+err.Error(), http.StatusUnauthorized) // Or InternalServerError
			return
		}
		if editor == nil { // Should be caught by err != nil if GetEditorByFirebaseUID returns sql.ErrNoRows
			handler.JSONError(w, "Editor not found for Firebase UID.", http.StatusUnauthorized)
			return
		}
		editorID := editor.ID // Editor's database ID

		// 3. Authorize: Check if the authenticated editor owns this newsletter
		// We need a method in newsletterRepo or newsletterService like GetNewsletterByIDAndEditorID
		// Let's assume newsletterRepo.GetNewsletterByIDAndEditorID(newsletterID, editorID) exists
		_, err = newsletterRepo.GetNewsletterByIDAndEditorID(newsletterID, editorID)
		if err != nil {
			// This could be because the newsletter doesn't exist or the editor doesn't own it, or a DB error.
			// service.ErrNewsletterNotFound is a good generic error here, or a specific auth error.
			// For simplicity, if it's not found for this editor, it's effectively a "not found" or "forbidden".
			handler.JSONError(w, "Newsletter not found or access denied.", http.StatusNotFound) // Or http.StatusForbidden
			return
		}

		// 4. Fetch active subscribers
		subscribers, err := subscriberService.GetActiveSubscribersForNewsletter(ctx, newsletterID)
		if err != nil {
			// service.ErrNewsletterNotFound could also be returned by GetActiveSubscribersForNewsletter
			// if the service layer itself checks for newsletter existence.
			if err == service.ErrNewsletterNotFound {
				handler.JSONError(w, "Newsletter not found when fetching subscribers.", http.StatusNotFound)
			} else {
				handler.JSONError(w, "Failed to get subscribers: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if subscribers == nil {
			subscribers = []models.Subscriber{} // Return empty list instead of null if no subscribers
		}
		
		handler.JSONResponse(w, subscribers, http.StatusOK)
	}
}
