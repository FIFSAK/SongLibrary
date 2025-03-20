package handlers

import (
	"SongLibrary/pkg/logger"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"SongLibrary/internal/models"
	"github.com/gin-gonic/gin"
)

var ExternalAPIURL = "http://localhost:8081"

// GetSongsHandler godoc
// @Summary      Get songs
// @Description  Get list of songs with filtering and pagination
// @Tags         songs
// @Param        id           query     int    false  "Song ID"
// @Param        group        query     string false  "Group name"
// @Param        song         query     string false  "Song name"
// @Param        releaseDate  query     string false  "Release date" format(date)
// @Param        text         query     string false  "Text fragment"
// @Param        page         query     int    false  "Page number"
// @Param        limit        query     int    false  "Items per page"
// @Success      200  {array}  models.Song
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /songs [get]
func GetSongsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Handling GET /songs request")

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid page parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page"})
			return
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid limit parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}

		id := 0
		idStr := c.Query("id")
		if idStr != "" {
			id, err = strconv.Atoi(idStr)
			if err != nil {
				logger.Log.WithError(err).Debug("Invalid ID parameter")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
		}

		releaseDateStr := c.Query("releaseDate")
		var releaseDate time.Time
		if releaseDateStr != "" {
			parsedDate, err := parseDateFlexible(releaseDateStr)
			if err != nil {
				logger.Log.WithError(err).Debug("Invalid releaseDate format")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid releaseDate format"})
			} else {
				releaseDate = parsedDate
			}
		}

		filter := models.SongFilter{
			ID:          uint(id),
			GroupName:   c.Query("group"),
			SongName:    c.Query("song"),
			Text:        c.Query("text"),
			ReleaseDate: releaseDate,
			Page:        page,
			Limit:       limit,
		}

		logger.Log.Debugf("Filter parameters: %+v", filter)

		songs, err := models.GetSongs(db, filter)
		if err != nil {
			logger.Log.WithError(err).Error("Failed to fetch songs from database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Log.Infof("Found %d songs matching filter", len(songs))
		c.JSON(http.StatusOK, songs)
	}
}

// GetSongVersesHandler godoc
// @Summary      Get song verses
// @Description  Get paginated verses of a song by its ID (split by paragraphs)
// @Tags         songs
// @Param        id     path      int     true  "Song ID"
// @Param        page   query     int     false "Page number (default 1)"
// @Param        limit  query     int     false "Verses per page (default 3)"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]interface{}
// @Failure      404    {object}  map[string]interface{}
// @Router       /songs/{id}/verses [get]
func GetSongVersesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Handling GET /songs/:id/verses request")

		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid ID parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid page parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page"})
			return
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "3"))
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid limit parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}

		logger.Log.Debugf("Request params — ID: %d, Page: %d, Limit: %d", id, page, limit)

		verses, err := models.GetSongVerses(db, uint(id), page, limit)
		if err != nil {
			logger.Log.WithError(err).Infof("Song with ID %d not found", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
			return
		}

		logger.Log.Infof("Returning %d verses for song ID %d", len(verses), id)
		c.JSON(http.StatusOK, gin.H{"verses": verses})
	}
}

// CreateSongHandler godoc
// @Summary      Add song
// @Description  Add song using external API enrichment
// @Tags         songs
// @Accept       json
// @Produce      json
// @Param        song  body  models.CreateSongInput  true  "Group and Song"
// @Success      201   {object}  models.Song
// @Failure      400   {object}  map[string]interface{}
// @Failure      502   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /songs [post]
func CreateSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Handling POST /songs request")

		var input models.CreateSongInput

		if err := c.ShouldBindJSON(&input); err != nil {
			logger.Log.WithError(err).Debug("Invalid JSON input")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Log.Infof("Received new song input: Group=%s, Song=%s", input.Group, input.Song)

		apiURL := fmt.Sprintf("%s/info?group=%s&song=%s",
			ExternalAPIURL, url.QueryEscape(input.Group), url.QueryEscape(input.Song))

		logger.Log.Debugf("Requesting external API: %s", apiURL)

		resp, err := http.Get(apiURL)
		if err != nil {
			logger.Log.WithError(err).Error("Failed to contact external API")
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to contact external API"})
			return
		}
		defer resp.Body.Close()

		logger.Log.Debugf("External API response status: %d", resp.StatusCode)

		if resp.StatusCode != http.StatusOK {
			logger.Log.Warnf("External API returned non-200 status: %d", resp.StatusCode)
			c.JSON(http.StatusBadGateway, gin.H{"error": "External API returned non-200 status"})
			return
		}

		var externalData struct {
			ReleaseDate string `json:"releaseDate"`
			Text        string `json:"text"`
			Link        string `json:"link"`
		}

		if err = json.NewDecoder(resp.Body).Decode(&externalData); err != nil {
			logger.Log.WithError(err).Error("Failed to parse external API response")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse external API response"})
			return
		}

		logger.Log.Debugf("External API data: %+v", externalData)

		releaseDate, err := parseDateFlexible(externalData.ReleaseDate)
		if err != nil {
			logger.Log.WithError(err).Error("Invalid date format from external API")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format from external API"})
			return
		}

		newSong := models.Song{
			GroupName:   input.Group,
			SongName:    input.Song,
			ReleaseDate: releaseDate,
			Text:        externalData.Text,
			Link:        externalData.Link,
		}

		if err = models.CreateSong(db, &newSong); err != nil {
			logger.Log.WithError(err).Error("Failed to save song in database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Log.Infof("Song successfully created: Group=%s, Song=%s", newSong.GroupName, newSong.SongName)
		c.JSON(http.StatusCreated, newSong)
	}
}

// UpdateSongHandler godoc
// @Summary      Update song
// @Description  Update an existing song by its ID
// @Tags         songs
// @Accept       json
// @Produce      json
// @Param        id    path   int            true  "Song ID"
// @Param        song  body   models.UpdateSongInput    true  "Updated song object"
// @Success      200   {object}  models.Song
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /songs/{id} [put]
func UpdateSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Handling PUT /songs/:id request")

		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid ID parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var updateSong models.UpdateSongInput
		if err = c.ShouldBindJSON(&updateSong); err != nil {
			logger.Log.WithError(err).Debug("Invalid JSON input")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ID := uint(id)
		parsedDate, err := parseDateFlexible(updateSong.ReleaseDate)
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid date format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		logger.Log.Debugf("Updating song with ID %d: %+v", id, updateSong)

		song := models.Song{
			ID:          ID,
			GroupName:   updateSong.GroupName,
			SongName:    updateSong.SongName,
			ReleaseDate: parsedDate,
			Text:        updateSong.Text,
			Link:        updateSong.Link,
		}

		if err = models.UpdateSong(db, song); err != nil {
			logger.Log.WithError(err).Errorf("Failed to update song ID %d", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Log.Infof("Song updated successfully: ID %d", id)
		c.JSON(http.StatusOK, song)
	}
}

// DeleteSongHandler godoc
// @Summary      Delete song
// @Description  Delete a song by its ID
// @Tags         songs
// @Produce      json
// @Param        id   path   int   true  "Song ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /songs/{id} [delete]
func DeleteSongHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Handling DELETE /songs/:id request")

		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			logger.Log.WithError(err).Debug("Invalid ID parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		logger.Log.Debugf("Deleting song with ID %d", id)

		if err = models.DeleteSong(db, uint(id)); err != nil {
			logger.Log.WithError(err).Errorf("Failed to delete song ID %d", id)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Log.Infof("Song deleted successfully: ID %d", id)
		c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
	}
}

func parseDateFlexible(dateStr string) (time.Time, error) {
	formats := []string{"2006.01.02", "2006-01-02", time.RFC3339}
	var err error
	for _, layout := range formats {
		var t time.Time
		t, err = time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", dateStr)
}
