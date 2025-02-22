package song_service

import (
	"context"
	"errors"
	"fmt"
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
	GetSongText(songID string, groupID uint, offset, limit int) ([]string, error)
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
	var allSongs []map[string]interface{}

	// Проходим по всем шардам и собираем данные
	for i := 1; i < 4; i++ {
		var songs []map[string]interface{}
		tableName := fmt.Sprintf("songs_%d", i)

		query := s.db.Table(tableName)

		// Применяем фильтры
		for field, value := range filters {
			query = query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
		}

		// Запрос данных из текущего шарда
		if err := query.Find(&songs).Error; err != nil {
			return nil, err
		}

		// Добавляем данные в общий массив
		allSongs = append(allSongs, songs...)
	}

	// Применяем offset и limit к общему списку
	start := offset
	if start > len(allSongs) {
		start = len(allSongs) // Если offset больше длины слайса, начинаем с конца
	}

	end := offset + limit
	if end > len(allSongs) {
		end = len(allSongs) // Если end выходит за границы слайса, обрезаем до конца
	}

	// Возвращаем срез с учетом offset и limit
	return allSongs[start:end], nil
}

func (s *SongService) GetSongText(songID string, groupID uint, offset, limit int) ([]string, error) {
	songText := new(dto.SongText)

	// Определяем шард, в котором хранится песня
	shardIndex := groupID % 4
	tableName := fmt.Sprintf("songs_%d", shardIndex)
	s.logger.Infof("table name: ", tableName)

	// Ищем текст песни в нужном шарде
	if err := s.db.Table(tableName).Select("text").Where("id = ? AND group_id = ?", songID, groupID).Scan(&songText).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("song not found")
		}
		return nil, err
	}

	s.logger.Infof("song text: ", songText.Text)

	// Проверяем, есть ли текст у песни
	if songText.Text == "" {
		return nil, errors.New("no text available for this song")
	}

	// Разделяем текст песни на массив строк (куплеты и припевы)
	sections := splitSongIntoSections(songText.Text)

	// Проверяем, есть ли строки текста
	if len(sections) == 0 {
		return nil, errors.New("no lyrics found for this song")
	}

	// Проверяем, не выходит ли offset за границы
	start := (offset - 1) * limit
	if start >= len(sections) {
		return []string{}, nil // Возвращаем пустой массив вместо ошибки
	}

	// Рассчитываем конечный индекс секций
	end := start + limit
	if end > len(sections) {
		end = len(sections) // Корректируем, если выходит за границы
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
	if err := s.db.Delete(&dto.Song{}, songID).Error; err != nil {
		return err
	}
	return nil
}

func (s *SongService) UpdateSong(songID string, data map[string]interface{}) error {
	if err := s.db.Model(&dto.Song{}).Where("id = ?", songID).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (s *SongService) AddSong(group, song string) error {
	// externalAPI := fmt.Sprintf(s.config.ExternalApi, group, song)
	// resp, err := http.Get(externalAPI)
	// if err != nil {
	// 	return errors.New("failed to call external API")
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return errors.New("external API error")
	// }

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return errors.New("failed to read API response body")
	// }

	// songDetails := new(dto.SongDetails)
	// if err := json.Unmarshal(body, &songDetails); err != nil {
	// 	return errors.New("failed to parse API response")
	// }

	var songDetails = dto.SongDetails{
		Group: "Muse3",
		Song:  "Supermassive Black Hole!!",
	}

	groupID, err := s.repo.CheckTable(context.Background(), songDetails.Group)
	if err != nil {
		return err
	}

	err = s.db.Create(&dto.Song{
		ID:      uuid.New(),
		GroupID: groupID,
		Song:    songDetails.Song,
	}).Error
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
