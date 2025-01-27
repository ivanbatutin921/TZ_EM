package song_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	song_dto "root/module/song/dto"
	"root/shared/logger"
	"strings"

	"gorm.io/gorm"
)

type ISongService interface {
	GetSongs(filters map[string]string,offset, limit int) ([]map[string]interface{}, error)
	GetSongText(songID string, offset, limit int) ([]string, error)
	DeleteSong(songID string) error
	UpdateSong(songID string, data map[string]interface{}) error
	AddSong(group, song string) error
}

type SongService struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewSongService(logger *logger.Logger, db *gorm.DB) ISongService {
	return &SongService{
		db:     db,
		logger: logger,
	}
}

func (s *SongService) GetSongs(filters map[string]string, offset, limit int) ([]map[string]interface{}, error) {
	var songs []map[string]interface{}

	// Создаем запрос с пагинацией
	query := s.db.Table("songs").Offset(offset).Limit(limit)

	// Применяем фильтры, если они указаны
	for field, value := range filters {
		// Поддержка частичного поиска (LIKE)
		query = query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	}

	// Выполняем запрос
	if err := query.Find(&songs).Error; err != nil {
		return nil, err
	}

	return songs, nil
}

func (s *SongService) GetSongText(songID string, offset, limit int) ([]string, error) {
	song := new(song_dto.SongText)
	if err := s.db.Table("songs").Select("text").Where("id = ?", songID).Scan(&song).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("song not found")
		}
		return nil, err
	}

	// Разделяем текст песни на массив строк (куплеты и припевы)
	sections := splitSongIntoSections(song.Text)

	// Проверяем корректность параметров страницы и лимита
	if offset < 1 || limit < 1 {
		return nil, errors.New("invalid page or limit")
	}

	// Рассчитываем диапазон (начальный и конечный индексы секций)
	start := (offset - 1) * limit
	end := start + limit

	// Проверяем выход за пределы массива
	if start >= len(sections) {
		return nil, errors.New("page out of range")
	}
	if end > len(sections) {
		end = len(sections) // Корректируем конечный индекс, если он превышает количество секций
	}

	// Возвращаем указанный диапазон секций
	return sections[start:end], nil
}

func splitSongIntoSections(songText string) []string {
	// Регулярное выражение для разделения по "Куплет" и "Припев"
	re := regexp.MustCompile(`(?i)(Куплет \d*:|Припев:)`)
	matches := re.FindAllStringIndex(songText, -1)

	var sections []string
	for i, match := range matches {
		start := match[0]
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(songText)
		}
		section := strings.TrimSpace(songText[start:end])
		if section != "" {
			sections = append(sections, section)
		}
	}

	return sections
}

func (s *SongService) DeleteSong(songID string) error {
	if err := s.db.Delete(&song_dto.Song{}, songID).Error; err != nil {
		return err
	}
	return nil
}

func (s *SongService) UpdateSong(songID string, data map[string]interface{}) error {
	if err := s.db.Model(&song_dto.Song{}).Where("id = ?", songID).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (s *SongService) AddSong(group, song string) error {
	apiURL := fmt.Sprintf("http://external-api.com/info?group=%s&song=%s", group, song)

	resp, err := http.Get(apiURL)
	if err != nil {
		return errors.New("failed to call external API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("external API error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("failed to read API response body")
	}

	songDetails := new(song_dto.SongDetails)
	if err := json.Unmarshal(body, &songDetails); err != nil {
		return errors.New("failed to parse API response")
	}

	newSong := song_dto.Song{
		Group:       group,
		Song:        song,
		ReleaseDate: songDetails.ReleaseDate,
		Text:        songDetails.Text,
		Link:        songDetails.Link,
	}

	if err := s.db.Create(&newSong).Error; err != nil {
		return err
	}

	return nil
}
