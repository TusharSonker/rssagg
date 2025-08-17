package database

import (
	"context"

	"github.com/google/uuid"
)

// GetPostsForUserPerFeed returns up to perFeedLimit newest posts for each feed the user follows.
func (q *Queries) GetPostsForUserPerFeed(ctx context.Context, userID uuid.UUID, perFeedLimit int32) ([]Post, error) {
	const getPostsPerFeed = `WITH user_feeds AS (
        SELECT feed_id FROM feed_follows WHERE user_id = $1
    ), ranked AS (
        SELECT p.*, ROW_NUMBER() OVER (PARTITION BY p.feed_id ORDER BY p.published_at DESC) AS rn
        FROM posts p
        JOIN user_feeds uf ON p.feed_id = uf.feed_id
    )
    SELECT id, created_at, updated_at, title, description, published_at, url, feed_id
    FROM ranked
    WHERE rn <= $2
    ORDER BY feed_id, published_at DESC`

	rows, err := q.db.QueryContext(ctx, getPostsPerFeed, userID, perFeedLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var i Post
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Description,
			&i.PublishedAt,
			&i.Url,
			&i.FeedID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
