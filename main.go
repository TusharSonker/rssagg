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

	// Mount all API routes under /projects/rssagg
	apiRouter := chi.NewRouter()

	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/err", handlerErr)
	apiRouter.Post("/users", apiCfg.handlerCreateUser)
	apiRouter.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	apiRouter.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	apiRouter.Get("/feeds", apiCfg.handlerGetFeeds)
	apiRouter.Post("/feed_follow", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	apiRouter.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	apiRouter.Delete("/feed_follow/{feedFollowId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))
	apiRouter.Get("/user_posts", apiCfg.middlewareAuth(apiCfg.handlerGetUserPosts))

	router.Mount("/projects/rssagg", apiRouter)

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
