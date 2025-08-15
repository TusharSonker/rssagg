# RSSAgg - RSS Aggregator

A full-stack RSS reader built with Go, PostgreSQL, and a modern web frontend.

## Features
- User authentication with API keys
- RSS feed management
- Background feed scraping
- Modern responsive UI
- Real-time post updates

## Tech Stack
- **Backend**: Go, Chi router, PostgreSQL
- **Frontend**: HTML/CSS/JS with Tailwind CSS
- **Database**: PostgreSQL with sqlc
- **Deployment**: Docker ready

## Quick Start
1. Clone the repository
2. Set up PostgreSQL database
3. Run migrations: `goose -dir sql/schema postgres "your-db-url" up`
4. Set environment variables (PORT, DB_URL)
5. Run: `go run .`
6. Open: `http://localhost:8080`

## API Endpoints
- `POST /v1/users` - Create user
- `POST /v1/feeds` - Create feed
- `GET /v1/feeds` - List feeds
- `POST /v1/feed_follow` - Follow feed
- `GET /v1/user_posts` - Get user posts

## License
MIT
EOF