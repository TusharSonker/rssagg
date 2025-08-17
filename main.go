package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TusharSonker/rssagg/internal/database"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {

	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}
	fmt.Println("Port:", portString)

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not found in the environment")
	}

	log.Printf("Connecting to database: %s", dbURL)
	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Can't connect to Database: ", err)
	}

	// Verify connection
	if err := conn.Ping(); err != nil {
		log.Fatal("Database ping failed: ", err)
	}

	log.Println("Database connection established successfully")
	dbQueries := database.New(conn)
	apiCfg := apiConfig{
		DB: dbQueries,
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Dedicated API under /projects/rssagg/api to isolate from SPA routing/caching
	router.Route("/projects/rssagg/api", func(r chi.Router) {
		// Logging middleware for API requests
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("API %s %s", r.Method, r.URL.Path)
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/healthz", handlerReadiness)
		r.Get("/err", handlerErr)
		r.Post("/users", apiCfg.handlerCreateUser)
		r.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
		r.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
		r.Get("/feeds", apiCfg.handlerGetFeeds)
		r.Post("/feed_follow", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
		r.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
		r.Delete("/feed_follow/{feedFollowId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))
		r.Get("/user_posts", apiCfg.middlewareAuth(apiCfg.handlerGetUserPosts))
		// Explicit 404
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			respondWithError(w, http.StatusNotFound, "not found")
		})
	})

	// Serve SPA under /projects/rssagg (no API routes here now)
	router.Get("/projects/rssagg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	})

	// Keep root serving index for local development without prefix (optional)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	})

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go startScraping(dbQueries, collectionConcurrency, collectionInterval)

	log.Printf("Server starting on port %v", portString)
	log.Fatal(srv.ListenAndServe())

	if err != nil {
		log.Fatal(err)
	}

}
