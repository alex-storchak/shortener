package handler

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/codec"
	"github.com/alex-storchak/shortener/internal/model"
)

// APIUserURLsProcessor defines the interface for processing user URL management operations.
// It provides methods for retrieving user's URLs and batch deletion of URLs.
type APIUserURLsProcessor interface {
	ProcessGet(ctx context.Context) (model.UserURLsGetResponse, error)
	ProcessDelete(ctx context.Context, shortIDs model.UserURLsDelRequest) error
}

// HandleGetUserURLs creates an HTTP handler for retrieving all URLs shortened by the authenticated user.
// It handles GET requests to '/api/user/urls' endpoint.
//
// The handler:
//   - Retrieves all URLs belonging to the authenticated user
//   - Returns appropriate HTTP status codes:
//   - 200 OK with model.UserURLsGetResponse when URLs are found
//   - 204 No Content when user has no URLs
//   - 500 Internal Server Error for processing failures
//
// Parameters:
//   - p: Processor implementing the user URLs retrieval logic
//   - l: Logger for logging operations
//
// Returns:
//   - HTTP handler function for the get user URLs endpoint
func HandleGetUserURLs(p APIUserURLsProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respItems, err := p.ProcessGet(r.Context())
		if err != nil {
			l.Error("error getting user urls", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if len(respItems) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err = codec.EasyJSONEncode(w, http.StatusOK, &respItems); err != nil {
			l.Error("encode json response", zap.Error(err))
			return
		}
	}
}

// HandleDeleteUserURLs creates an HTTP handler for batch deletion of user's URLs.
// It handles DELETE requests to '/api/user/urls' endpoint
// with JSON body containing short IDs to delete.
//
// The handler:
//   - Processes the batch deletion request asynchronously
//   - Returns appropriate HTTP status codes:
//   - 202 Accepted for successful request acceptance
//   - 400 Bad Request for malformed JSON
//   - 500 Internal Server Error for processing failures
//
// Note: Deletion is processed asynchronously,
// the handler returns immediately after accepting the request.
//
// Parameters:
//   - p: Processor implementing the batch deletion logic
//   - l: Logger for logging operations
//
// Returns:
//   - HTTP handler function for the delete user URLs endpoint
func HandleDeleteUserURLs(p APIUserURLsProcessor, l *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var shortIDs model.UserURLsDelRequest
		if err := codec.EasyJSONDecode(r, &shortIDs); err != nil {
			l.Debug("decode json request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := p.ProcessDelete(r.Context(), shortIDs); err != nil {
			l.Error("error deleting user urls", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
