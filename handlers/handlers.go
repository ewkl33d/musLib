package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"musLib/db"
	"musLib/env"
	"musLib/logger"
	"net/http"
	"strconv"
	"strings"
)

type SongHandler struct {
	DB *gorm.DB
}

// GetSongs godoc
// @Summary Get songs
// @Description Get a list of songs with optional filtering and pagination
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Filter by group name"
// @Param song query string false "Filter by song name"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {array} db.Song
// @Router /songs [get]
func (h *SongHandler) GetSongs(c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("Handling GET /songs request")

	var songs []db.Song
	query := h.DB

	// Фильтрация по полям
	if group := c.Query("group"); group != "" {
		logger.Debug("Filtering by group: " + group)
		query = query.Where("group_name = ?", group)
	}
	if song := c.Query("song"); song != "" {
		logger.Debug("Filtering by song: " + song)
		query = query.Where("song_name = ?", song)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	logger.Debug(fmt.Sprintf("Fetching songs with offset: %d, limit: %d", offset, limit))
	query.Offset(offset).Limit(limit).Find(&songs)

	logger.Info("Returning songs")
	c.JSON(http.StatusOK, songs)
}

// GetSongText godoc
// @Summary Get song text
// @Description Get the text of a specific song by ID with optional pagination
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} map[string]interface{}
// @Router /songs/{id}/text [get]
func (h *SongHandler) GetSongText(c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("Handling GET /songs/:id/text request")

	var song db.Song
	if err := h.DB.Where("id = ?", c.Param("id")).First(&song).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	verses := strings.Split(song.Text, "\n\n")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1"))
	offset := (page - 1) * limit

	if offset >= len(verses) {
		c.JSON(http.StatusOK, gin.H{"text": ""})
		return
	}

	end := offset + limit
	if end > len(verses) {
		end = len(verses)
	}

	logger.Info("Returning song text")
	c.JSON(http.StatusOK, gin.H{"text": verses[offset:end]})
}

// DeleteSong godoc
// @Summary Delete song
// @Description Delete a specific song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Success 200 {object} map[string]interface{}
// @Router /songs/{id} [delete]
func (h *SongHandler) DeleteSong(c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("Handling DELETE /songs/:id request")

	if err := h.DB.Where("id = ?", c.Param("id")).Delete(&db.Song{}).Error; err != nil {
		logger.Error("Song not found: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	logger.Info("Song deleted")
	c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
}

// UpdateSong godoc
// @Summary Update song
// @Description Update a specific song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body db.Song true "Song object"
// @Success 200 {object} db.Song
// @Router /songs/{id} [put]
func (h *SongHandler) UpdateSong(c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("Handling PUT /songs/:id request")

	var song db.Song
	if err := h.DB.Where("id = ?", c.Param("id")).First(&song).Error; err != nil {
		logger.Error("Song not found: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	if err := c.BindJSON(&song); err != nil {
		logger.Error("Invalid request: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	h.DB.Save(&song)
	logger.Info("Song updated")
	c.JSON(http.StatusOK, song)
}

// AddSong godoc
// @Summary Add song
// @Description Add a new song
// @Tags songs
// @Accept json
// @Produce json
// @Param song body db.Song true "Song object"
// @Success 201 {object} db.Song
// @Router /songs/add [post]
func (h *SongHandler) AddSong(c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("Handling POST /songs/add request")

	var input struct {
		Group string `json:"group"`
		Song  string `json:"song"`
	}

	if err := c.BindJSON(&input); err != nil {
		logger.Error("Invalid JSON: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	song := db.Song{
		GroupName: input.Group,
		SongName:  input.Song,
	}

	h.DB.Create(&song)
	logger.Info("Song added")
	c.JSON(http.StatusCreated, song)
}

func requestToApi(song *db.Song, c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("Requesting song details from API")

	// Запрос к Апи
	apiUrl := env.GetEnv("API_URL", "")
	logger.Debug("API URL: " + apiUrl)

	// Запрос к API
	apiURL := fmt.Sprintf(apiUrl+"?group=%s&song=%s", song.GroupName, song.SongName)
	logger.Debug("API request URL: " + apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		logger.Error("Failed to fetch song details: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch song details"})
		return
	}
	defer resp.Body.Close()

	// Ответ от API
	var songDetails map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		logger.Error("Failed to decode song details: " + err.Error())
		song.ReleaseDate = "foo"
		song.Text = "bar"
		song.Link = "foobar"
	} else {
		song.ReleaseDate = songDetails["releaseDate"]
		song.Text = songDetails["text"]
		song.Link = songDetails["link"]
	}

	logger.Info("Song details gets updated")
}
