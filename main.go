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

	// API routes at both root level and under /projects/rssagg for the frontend
	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerErr)
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	v1Router.Post("/feed_follow", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follow/{feedFollowId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))
	v1Router.Get("/user_posts", apiCfg.middlewareAuth(apiCfg.handlerGetUserPosts))

	// Mount the API routes both at root and under /projects/rssagg
	router.Mount("/", v1Router)
	router.Mount("/projects/rssagg", v1Router)

	// Serve the frontend (single-page app) from ./public at /projects/rssagg and all subpaths
	router.Get("/projects/rssagg*", func(w http.ResponseWriter, r *http.Request) {
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
