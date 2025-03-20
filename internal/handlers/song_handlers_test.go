package handlers

import (
	"SongLibrary/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.Song{})
	require.NoError(t, err)
	return db
}

func TestCreateSongHandler(t *testing.T) {
	db := setupTestDB(t)

	mockExternalAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"releaseDate": "2010-01-01",
			"text":        "Test Lyrics",
			"link":        "https://example.com/song",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockExternalAPI.Close()

	ExternalAPIURL = mockExternalAPI.URL

	router := gin.Default()
	router.POST("/songs", CreateSongHandler(db))

	requestBody := `{"group": "Test Group", "song": "Test Song"}`
	req, _ := http.NewRequest("POST", "/songs", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdSong models.Song
	err := json.Unmarshal(w.Body.Bytes(), &createdSong)
	require.NoError(t, err)
	assert.Equal(t, "Test Group", createdSong.GroupName)
	assert.Equal(t, "Test Song", createdSong.SongName)
	assert.Equal(t, "Test Lyrics", createdSong.Text)
}

func TestUpdateSongHandler(t *testing.T) {
	db := setupTestDB(t)

	song := models.Song{
		GroupName:   "Muse",
		SongName:    "OldSong",
		ReleaseDate: time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC),
		Text:        "Old lyrics",
		Link:        "https://youtube.com/old",
	}
	db.Create(&song)
	fmt.Println(song)

	router := gin.Default()
	router.PUT("/songs/:id", UpdateSongHandler(db))

	update := map[string]string{
		"group_name":   "Muse",
		"song_name":    "NewSong",
		"release_date": "2009-09-07",
		"text":         "New lyrics",
		"link":         "https://youtube.com/new",
	}
	body, _ := json.Marshal(update)
	url := "/songs/" + strconv.Itoa(int(song.ID))
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Song
	err := json.Unmarshal(w.Body.Bytes(), &updated)
	assert.NoError(t, err)
	assert.Equal(t, "NewSong", updated.SongName)
	assert.Equal(t, "New lyrics", updated.Text)
}

func TestDeleteSongHandler(t *testing.T) {
	db := setupTestDB(t)

	song := models.Song{
		GroupName:   "Muse",
		SongName:    "DeleteMe",
		ReleaseDate: time.Now(),
		Text:        "Lyrics",
		Link:        "https://link",
	}
	db.Create(&song)

	router := gin.Default()
	router.DELETE("/songs/:id", DeleteSongHandler(db))

	url := "/songs/" + strconv.Itoa(int(song.ID))
	req, _ := http.NewRequest("DELETE", url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "Song deleted", result["message"])
}

func TestGetSongVersesHandler(t *testing.T) {
	db := setupTestDB(t)

	song := models.Song{
		GroupName:   "Muse",
		SongName:    "Verses",
		ReleaseDate: time.Now(),
		Text:        "Line1\n\nLine2\n\nLine3\n\nLine4",
		Link:        "https://link",
	}
	db.Create(&song)

	router := gin.Default()
	router.GET("/songs/:id/verses", GetSongVersesHandler(db))

	url := "/songs/" + strconv.Itoa(int(song.ID)) + "/verses?page=1&limit=2"
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response["verses"], 2)
	assert.Equal(t, "Line1", response["verses"][0])
}
