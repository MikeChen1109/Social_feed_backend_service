package main

import (
	"feed-service/initializers"
	"feed-service/models"
	"log"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	db := initializers.ConnectToDatabase()
	// db.Migrator().DropTable(&models.Comment{}, &models.Feed{})
	if err := db.AutoMigrate(&models.Feed{}, &models.Comment{}); err != nil {
		log.Fatal(err)
	}
	for _, query := range []string{
		"CREATE INDEX IF NOT EXISTS idx_feeds_cursor ON feeds (created_at DESC, id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_comments_cursor ON comments (feed_id, created_at DESC, id DESC) WHERE deleted_at IS NULL",
	} {
		if err := db.Exec(query).Error; err != nil {
			log.Fatal(err)
		}
	}
}
