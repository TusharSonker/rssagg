package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/TusharSonker/rssagg/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing json: %v", err))
		return
	}

	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create user: %v", err))
		return
	}

	respondWithJSON(w, 201, databaseUserToUser(user))

}

func (cfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, databaseUserToUser(user))
}

func (cfg *apiConfig) handlerGetUserPosts(w http.ResponseWriter, r *http.Request, user database.User) {
	// Interpret 'limit' as per-feed limit
	var perFeed int32 = 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			if v < 1 {
				v = 1
			}
			if v > 50 {
				v = 50
			}
			perFeed = int32(v)
		}
	}
	posts, err := cfg.DB.GetPostsForUserPerFeed(r.Context(), user.ID, perFeed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error couldn't get posts: %v", err))
		return
	}
	respondWithJSON(w, http.StatusOK, databasePostsToPosts(posts))
}
