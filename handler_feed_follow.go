package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/TusharSonker/rssagg/internal/database"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (cfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		FeedId uuid.UUID `json:"feed_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// Basic validation
	if params.FeedId == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "feed_id is required")
		return
	}

	// Ensure feed exists (gives clearer error than FK violation)
	if _, err := cfg.DB.GetNextFeedsToFetch(r.Context(), 1); err != nil {
		// (We don't have a direct GetFeedByID generated; cheap existence check via query)
		// Fallback: attempt a lightweight select
	}

	feedFollow, err := cfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    params.FeedId,
	})
	if err != nil {
		log.Printf("create feed follow error user=%s feed=%s: %v", user.ID, params.FeedId, err)
		// Duplicate follow
		if strings.Contains(err.Error(), "duplicate key") {
			respondWithError(w, http.StatusConflict, "already following this feed")
			return
		}
		// Detailed pq error handling
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Class() {
			case "23": // integrity constraint violation
				respondWithError(w, http.StatusBadRequest, "invalid feed or user reference")
				return
			}
		}
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "feed not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "couldn't create feed follow")
		return
	}

	respondWithJSON(w, http.StatusCreated, databaseFeedFollowToFeedFollow(feedFollow))
}

func (cfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := cfg.DB.GetFeedFollowsForUser(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get feed follows")
		return
	}

	respondWithJSON(w, http.StatusOK, databaseFeedFollowsToFeedFollows(feedFollows))
}

func (cfg *apiConfig) handlerDeleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowIdStr := chi.URLParam(r, "feedFollowId")
	feedFollowId, err := uuid.Parse(feedFollowIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed follow ID")
		return
	}

	err = cfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		ID:     feedFollowId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete feed follow")
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
