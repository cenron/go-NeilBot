package storage

import (
	"log"
	"time"
)

// SaveBootyMessage save the booty message to the database
func (s *Storage) SaveBootyMessage(messageID, channelID, guildID string, imageID int64) (int64, error) {
	query := "INSERT INTO booty_message (message_id, channel_id, guild_id, image_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?);"
	r, err := s.DB.Exec(query, messageID, channelID, guildID, imageID, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error inserting booty message: %v", err)
		return 0, err
	}
	return r.LastInsertId()
}

// SaveBootyImage save the booty image to the database
func (s *Storage) SaveBootyImage(name, mimeType, hash string) (int64, error) {
	query := "INSERT INTO booty_image (name, mime_type, hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?);"
	r, err := s.DB.Exec(query, name, mimeType, hash, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error inserting booty image: %v", err)
		return 0, err
	}
	return r.LastInsertId()
}

// AddBootyLike add a like to the booty message
func (s *Storage) AddBootyLike(messageID string) (int64, error) {
	// Need to get the image_id from the booty_message, then update the booty_image
	
	query := "UPDATE booty_message SET likes = likes + 1 WHERE message_id = ?;"
	r, err := s.DB.Exec(query, messageID)
	if err != nil {
		log.Printf("Error updating booty message: %v", err)
		return 0, err
	}

	return r.LastInsertId()
}
