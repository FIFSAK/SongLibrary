package models

import (
	"SongLibrary/pkg/logger"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Song struct {
	ID          uint      `gorm:"primaryKey"`
	GroupName   string    `gorm:"not null"`
	SongName    string    `gorm:"not null"`
	ReleaseDate time.Time `gorm:"not null"`
	Text        string    `gorm:"not null"`
	Link        string    `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SongFilter struct {
	ID          uint
	GroupName   string
	SongName    string
	ReleaseDate time.Time
	Text        string
	Page        int
	Limit       int
}

type CreateSongInput struct {
	Group string `json:"group" binding:"required" example:"Muse"`
	Song  string `json:"song" binding:"required" example:"Supermassive Black Hole"`
}

func GetSongs(db *gorm.DB, filter SongFilter) ([]Song, error) {
	var songs []Song
	query := db.Model(&Song{})

	logger.Log.Debug("Building query for GetSongs")

	if filter.ID != 0 {
		query = query.Where("id = ?", filter.ID)
		logger.Log.Debugf("Filter: ID = %d", filter.ID)
	}
	if filter.GroupName != "" {
		query = query.Where("group_name ILIKE ?", "%"+filter.GroupName+"%")
		logger.Log.Debugf("Filter: GroupName ILIKE '%%%s%%'", filter.GroupName)
	}
	if filter.SongName != "" {
		query = query.Where("song_name ILIKE ?", "%"+filter.SongName+"%")
		logger.Log.Debugf("Filter: SongName ILIKE '%%%s%%'", filter.SongName)
	}
	if !filter.ReleaseDate.IsZero() {
		query = query.Where("release_date = ?", filter.ReleaseDate)
		logger.Log.Debugf("Filter: ReleaseDate = %s", filter.ReleaseDate.Format("2006-01-02"))
	}
	if filter.Text != "" {
		query = query.Where("text ILIKE ?", "%"+filter.Text+"%")
		logger.Log.Debugf("Filter: Text ILIKE '%%%s%%'", filter.Text)
	}

	if filter.Limit == 0 {
		filter.Limit = 10
		logger.Log.Debug("No limit provided, defaulting to 10")
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit
	logger.Log.Debugf("Pagination: Page=%d, Limit=%d, Offset=%d", filter.Page, filter.Limit, offset)

	err := query.Limit(filter.Limit).Offset(offset).Find(&songs).Error
	if err != nil {
		logger.Log.WithError(err).Error("Failed to fetch songs from database")
	} else {
		logger.Log.Infof("Fetched %d song(s) from database", len(songs))
	}

	return songs, err
}

func GetSongVerses(db *gorm.DB, id uint, page, limit int) ([]string, error) {
	logger.Log.Debugf("Fetching song with ID: %d for verses", id)

	var song Song
	err := db.First(&song, id).Error
	if err != nil {
		logger.Log.WithError(err).Errorf("Failed to fetch song with ID %d", id)
		return nil, err
	}

	verses := strings.Split(song.Text, `\n\n`)
	fmt.Println(verses)
	totalVerses := len(verses)
	logger.Log.Debugf("Song ID %d has %d verses", id, totalVerses)

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 3
	}

	start := (page - 1) * limit
	end := start + limit

	if start > totalVerses {
		logger.Log.Infof("Pagination out of range: start=%d > total=%d", start, totalVerses)
		return []string{}, nil
	}
	if end > totalVerses {
		end = totalVerses
	}

	selectedVerses := verses[start:end]
	logger.Log.Infof("Returning verses %d to %d for song ID %d", start+1, end, id)
	return selectedVerses, nil
}

func CreateSong(db *gorm.DB, song *Song) error {
	logger.Log.Infof("Creating song: Group=%s, Song=%s", song.GroupName, song.SongName)

	err := db.Create(&song).Error
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create song in database")
	} else {
		logger.Log.Infof("Song created successfully: ID=%d", song.ID)
	}

	return err
}

func UpdateSong(db *gorm.DB, updatedSong Song) error {
	logger.Log.Debugf("Attempting to update song with ID=%d", updatedSong.ID)

	var existing Song
	err := db.First(&existing, updatedSong.ID).Error
	if err != nil {
		logger.Log.WithError(err).Errorf("Song with ID=%d not found for update", updatedSong.ID)
		return err
	}

	existing.GroupName = updatedSong.GroupName
	existing.SongName = updatedSong.SongName
	//existing.ReleaseDate = updatedSong.ReleaseDate
	existing.Text = updatedSong.Text
	existing.Link = updatedSong.Link

	err = db.Save(&existing).Error
	if err != nil {
		logger.Log.WithError(err).Errorf("Failed to update song ID=%d", updatedSong.ID)
	} else {
		logger.Log.Infof("Song updated successfully: ID=%d", updatedSong.ID)
	}

	return err
}

func DeleteSong(db *gorm.DB, id uint) error {
	logger.Log.Debugf("Attempting to delete song with ID=%d", id)

	err := db.Delete(&Song{}, id).Error
	if err != nil {
		logger.Log.WithError(err).Errorf("Failed to delete song ID=%d", id)
	} else {
		logger.Log.Infof("Song deleted successfully: ID=%d", id)
	}

	return err
}
