package song_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"root/config"
	dto "root/module/song/dto"
	song_repository "root/module/song/repository"
	"root/shared/logger"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ISongService interface {
	GetSongs(filters map[string]string, offset, limit int) ([]map[string]interface{}, error)
	GetSongText(songID string, offset, limit int) ([]string, error)
	DeleteSong(songID string) error
	UpdateSong(songID string, data map[string]interface{}) error
	AddSong(group, song string) error
}

type SongService struct {
	repo   song_repository.ISongRepository
	logger *logger.Logger
	config *config.Config
	db     *gorm.DB
}

func NewSongService(logger *logger.Logger, config *config.Config, db *gorm.DB, repo song_repository.ISongRepository) ISongService {
	return &SongService{
		logger: logger,
		config: config,
		db:     db,
		repo:   repo,
	}
}

func (s *SongService) GetSongs(filters map[string]string, offset, limit int) ([]map[string]interface{}, error) {
	s.logger.Info("GetSongs: started")
	defer s.logger.Info("GetSongs: completed")

	var allSongs []map[string]interface{}

	for i := 1; i < 4; i++ {
		tableName := fmt.Sprintf("songs_%d", i)
		s.logger.Infof("GetSongs: checking table %s", tableName)

		var songs []map[string]interface{}
		query := s.db.Table(tableName)

		for field, value := range filters {
			field = strings.ToLower(field)
			escapedField := fmt.Sprintf(`"%s"`, field)
			query = query.Where(fmt.Sprintf("%s LIKE ?", escapedField), "%"+value+"%")
		}

		s.logger.Infof("GetSongs: executing query for table %s", tableName)
		if err := query.Find(&songs).Error; err != nil {
			s.logger.Errorf("GetSongs: error fetching songs from table %s: %v", tableName, err)
			return nil, err
		}

		allSongs = append(allSongs, songs...)
		s.logger.Infof("GetSongs: found %d songs in table %s", len(songs), tableName)
	}

	start := offset
	if start > len(allSongs) {
		start = len(allSongs)
	}
	end := offset + limit
	if end > len(allSongs) {
		end = len(allSongs)
	}

	s.logger.Infof("GetSongs: returning %d songs (offset: %d, limit: %d)", len(allSongs[start:end]), offset, limit)
	return allSongs[start:end], nil
}

func (s *SongService) GetSongText(songID string, offset, limit int) ([]string, error) {
	s.logger.Info("GetSongText: started")
	defer s.logger.Info("GetSongText: completed")

	songText := new(dto.SongText)

	for i := 0; i < 4; i++ {
		tableName := fmt.Sprintf("songs_%d", i)
		s.logger.Infof("GetSongText: checking table %s", tableName)

		if err := s.db.Table(tableName).Select("text").Where("id = ?", songID).Scan(&songText).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Infof("GetSongText: song not found in table %s", tableName)
				continue
			}
			s.logger.Errorf("GetSongText: error fetching song text from table %s: %v", tableName, err)
			return nil, err
		}

		if songText.Text != "" {
			s.logger.Infof("GetSongText: song text found in table %s", tableName)
			break
		}
	}

	if songText.Text == "" {
		s.logger.Error("GetSongText: song not found in any table")
		return nil, errors.New("song not found")
	}

	s.logger.Infof("GetSongText: song text found: %s", songText.Text)

	sections := splitSongIntoSections(songText.Text)
	s.logger.Infof("GetSongText: split song into %d sections", len(sections))

	start := (offset - 1) * limit
	if start >= len(sections) {
		s.logger.Infof("GetSongText: offset %d is out of bounds, returning empty slice", offset)
		return []string{}, nil
	}

	end := start + limit
	if end > len(sections) {
		end = len(sections)
	}

	s.logger.Infof("GetSongText: returning %d sections (offset: %d, limit: %d)", len(sections[start:end]), offset, limit)
	return sections[start:end], nil
}

func (s *SongService) DeleteSong(songID string) error {
	s.logger.Info("DeleteSong: started")
	defer s.logger.Info("DeleteSong: completed")

	for i := 1; i <= 4; i++ {
		tableName := fmt.Sprintf("songs_%d", i)
		s.logger.Infof("DeleteSong: checking table %s", tableName)

		result := s.db.Table(tableName).Where("id = ?", songID).Delete(&dto.Song{})
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				s.logger.Infof("DeleteSong: song not found in table %s", tableName)
				continue
			}
			s.logger.Errorf("DeleteSong: error deleting song from table %s: %v", tableName, result.Error)
			return result.Error
		}

		if result.RowsAffected > 0 {
			s.logger.Infof("DeleteSong: song deleted from table %s", tableName)
			return nil
		}
	}

	s.logger.Error("DeleteSong: song not found in any table")
	return errors.New("song not found")
}

func (s *SongService) UpdateSong(songID string, data map[string]interface{}) error {
	s.logger.Info("UpdateSong: started")
	defer s.logger.Info("UpdateSong: completed")

	for i := 1; i <= 4; i++ {
		tableName := fmt.Sprintf("songs_%d", i)
		s.logger.Infof("UpdateSong: checking table %s", tableName)

		result := s.db.Table(tableName).Where("id = ?", songID).Updates(data)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				s.logger.Infof("UpdateSong: song not found in table %s", tableName)
				continue
			}
			s.logger.Errorf("UpdateSong: error updating song in table %s: %v", tableName, result.Error)
			return result.Error
		}

		if result.RowsAffected > 0 {
			s.logger.Infof("UpdateSong: song updated in table %s", tableName)
			return nil
		}
	}

	s.logger.Error("UpdateSong: song not found in any table")
	return errors.New("song not found")
}

func (s *SongService) AddSong(group, song string) error {
	s.logger.Info("AddSong: started")
	defer s.logger.Info("AddSong: completed")

	externalAPI := fmt.Sprintf(s.config.ExternalApi, group, song)
	s.logger.Infof("AddSong: calling external API: %s", externalAPI)

	resp, err := http.Get(externalAPI)
	if err != nil {
		s.logger.Errorf("AddSong: failed to call external API: %v", err)
		return errors.New("failed to call external API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("AddSong: external API returned status code %d", resp.StatusCode)
		return errors.New("external API error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("AddSong: failed to read API response body: %v", err)
		return errors.New("failed to read API response body")
	}

	songDetails := new(dto.SongDetails)
	if err := json.Unmarshal(body, &songDetails); err != nil {
		s.logger.Errorf("AddSong: failed to parse API response: %v", err)
		return errors.New("failed to parse API response")
	}

	s.logger.Infof("AddSong: fetched song details: %+v", songDetails)

	groupID, err := s.repo.CheckTable(context.Background(), songDetails.Group)
	if err != nil {
		s.logger.Errorf("AddSong: failed to check group table: %v", err)
		return err
	}

	s.logger.Infof("AddSong: group ID: %d", groupID)

	err = s.db.Create(&dto.Song{
		ID:      uuid.New(),
		GroupID: groupID,
		Group:   songDetails.Group,
		Song:    songDetails.Song,
	}).Error
	if err != nil {
		s.logger.Errorf("AddSong: failed to create song: %v", err)
		return fmt.Errorf("failed to create song: %w", err)
	}

	s.logger.Info("AddSong: song created successfully")
	return nil
}

func splitSongIntoSections(songText string) []string {
	re := regexp.MustCompile(`(?i)\[?(Куплет \d+|Припев)\]?\s*\n`)
	matches := re.FindAllStringIndex(songText, -1)

	var sections []string
	start := 0

	for _, match := range matches {
		if start < match[0] {
			section := strings.TrimSpace(songText[start:match[0]])
			if section != "" {
				sections = append(sections, section)
			}
		}
		start = match[1]
	}

	if start < len(songText) {
		section := strings.TrimSpace(songText[start:])
		if section != "" {
			sections = append(sections, section)
		}
	}

	return sections
}
