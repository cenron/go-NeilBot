package storage

import (
	"log"
	"time"
)

type BootyImageEntity struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	MimeType  string    `db:"mime_type"`
	Hash      string    `db:"hash"`
	Likes     int       `db:"likes"`
	Dislikes  int       `db:"dislikes"`
	PostCount int       `db:"post_count"`
	Rarity    string    `db:"rarity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

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
func (s *Storage) SaveBootyImage(fileName string, rarity string, mimeType, hash string) (int64, error) {

	query := "INSERT INTO booty_image (name, mime_type, hash, post_count, rarity, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?);"
	r, err := s.DB.Exec(query, fileName, mimeType, hash, 1, rarity, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error inserting booty image: %v", err)
		return 0, err
	}
	return r.LastInsertId()
}

// AddBootyLike add a like to the booty message
func (s *Storage) AddBootyLike(messageID string) (int64, error) {
	imageID, err := s.GetBootyImageIdByMessageId(messageID)
	if err != nil {
		log.Printf("Error getting image_id: %v", err)
		return 0, err
	}

	query := "UPDATE booty_image SET likes = likes + 1 WHERE id = ?;"
	r, err := s.DB.Exec(query, imageID)
	if err != nil {
		log.Printf("Error updating booty message: %v", err)
		return 0, err
	}

	return r.LastInsertId()
}

// RemoveBootyLike remove a like to the booty message
func (s *Storage) RemoveBootyLike(messageID string) (int64, error) {
	imageID, err := s.GetBootyImageIdByMessageId(messageID)
	if err != nil {
		log.Printf("Error getting image_id: %v", err)
		return 0, err
	}

	query := "UPDATE booty_image SET likes = likes - 1 WHERE id = ?;"
	r, err := s.DB.Exec(query, imageID)
	if err != nil {
		log.Printf("Error updating booty message: %v", err)
		return 0, err
	}

	return r.LastInsertId()
}

// AddBootyDislike add a dislike to the booty message
func (s *Storage) AddBootyDislike(messageID string) (int64, error) {
	imageID, err := s.GetBootyImageIdByMessageId(messageID)
	if err != nil {
		log.Printf("Error getting image_id: %v", err)
		return 0, err
	}

	query := "UPDATE booty_image SET dislikes = dislikes + 1 WHERE id = ?;"
	r, err := s.DB.Exec(query, imageID)
	if err != nil {
		log.Printf("Error updating booty message: %v", err)
		return 0, err
	}

	return r.LastInsertId()
}

// RemoveBootyDislike remove a dislike to the booty message
func (s *Storage) RemoveBootyDislike(messageID string) (int64, error) {
	imageID, err := s.GetBootyImageIdByMessageId(messageID)
	if err != nil {
		log.Printf("Error getting image_id: %v", err)
		return 0, err
	}

	query := "UPDATE booty_image SET dislikes = dislikes - 1 WHERE id = ?;"
	r, err := s.DB.Exec(query, imageID)
	if err != nil {
		log.Printf("Error updating booty message: %v", err)
		return 0, err
	}

	return r.LastInsertId()
}

func (s *Storage) GetBootyImageId(hash string) (int64, error) {
	query := "SELECT id FROM booty_image WHERE hash =?;"
	var id int64
	err := s.DB.Get(&id, query, hash)
	if err != nil {
		log.Printf("Error getting image_id by hash: %v", err)
		return 0, err
	}

	return id, nil
}

func (s *Storage) GetBootyImage(hash string) (*BootyImageEntity, error) {
	query := "SELECT * FROM booty_image WHERE hash =?;"
	var bie = BootyImageEntity{}
	err := s.DB.QueryRowx(query, hash).StructScan(&bie)
	if err != nil {
		log.Printf("Error getting image by hash: %v", err)
		return nil, err
	}

	return &bie, nil
}

func (s *Storage) IncrementPostCount(imageID int64) {
	query := "UPDATE booty_image SET post_count = post_count + 1 WHERE id = ?"
	_, err := s.DB.Exec(query, imageID)
	if err != nil {
		log.Printf("Error incrementing post count: %v", err)
	}
	return
}

func (s *Storage) UpdateRarity(imageID int64, rarity string) {
	query := "UPDATE booty_image SET rarity = ? WHERE id = ?"
	_, err := s.DB.Exec(query, rarity, imageID)
	if err != nil {
		log.Printf("Error incrementing post count: %v", err)
	}
	return
}

// GetBootyImageIdByMessageId get the image id the message is associated with.
func (s *Storage) GetBootyImageIdByMessageId(messageID string) (int64, error) {
	// Need to get the image_id from the booty_message, then update the booty_image
	query := "SELECT image_id FROM booty_message WHERE message_id = ?;"
	var imageID int64
	err := s.DB.Get(&imageID, query, messageID)
	if err != nil {
		log.Printf("Error getting image_id: %v", err)
		return 0, err
	}

	return imageID, nil
}
