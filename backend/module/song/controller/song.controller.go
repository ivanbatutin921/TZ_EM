package song_controller

import (
	"log"
	song_service "root/module/song/service"
	"root/shared/logger"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ISongController interface {
	GetSongs(c *fiber.Ctx) error
	GetSongText(c *fiber.Ctx) error
	DeleteSong(c *fiber.Ctx) error
	UpdateSong(c *fiber.Ctx) error
	AddSong(c *fiber.Ctx) error
}

type SongController struct {
	logger      *logger.Logger
	
	songService song_service.ISongService
}

func NewSongController(logger *logger.Logger, songService song_service.ISongService) ISongController {
	return &SongController{
		logger:      logger,
		songService: songService,
	}
}

// GetSongs возвращает список песен с возможностью фильтрации и пагинации
// @Summary Получение списка песен
// @Description Возвращает список песен с поддержкой фильтрации и пагинации
// @Tags Песни
// @Accept json
// @Produce json
// @Param offset query int false "Смещение для пагинации" default(0)
// @Param limit query int false "Количество записей на странице" default(10)
// @Param group query string false "Фильтр по группе"
// @Param song query string false "Фильтр по названию песни"
// @Param release_date query string false "Фильтр по дате релиза"
// @Param text query string false "Фильтр по тексту песни"
// @Param link query string false "Фильтр по ссылке"
// @Success 200 {array} map[string]interface{} "Список песен"
// @Failure 400 {object} map[string]interface{} "Неверный запрос (например, некорректный offset или limit)"
// @Failure 500 {object} map[string]interface{} "Ошибка сервера"
// @Router /api/songs [get]
func (sc *SongController) GetSongs(c *fiber.Ctx) error {
	// Получение параметров пагинации из строки запроса
	offset, err := strconv.Atoi(c.Query("offset", "0")) // "0" — значение по умолчанию
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid offset number"})
	}

	limit, err := strconv.Atoi(c.Query("limit", "10")) // "10" — значение по умолчанию
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid limit number"})
	}

	//можно еще добавить получение песен с  пагинаций в опредленной группе

	// Получение всех параметров фильтрации из строки запроса
	filters := map[string]string{}
	if group := c.Query("group"); group != "" {
		filters["Group"] = group
	}
	if song := c.Query("song"); song != "" {
		filters["Song"] = song
	}
	if releaseDate := c.Query("release_date"); releaseDate != "" {
		filters["ReleaseDate"] = releaseDate
	}
	if text := c.Query("text"); text != "" {
		filters["Text"] = text
	}
	if link := c.Query("link"); link != "" {
		filters["Link"] = link
	}

	log.Printf("offset: %d, limit: %d, filters: %v", offset, limit, filters)

	// Вызов сервиса для получения данных с фильтрацией
	result, err := sc.songService.GetSongs(filters, offset, limit)
	if err != nil {
		sc.logger.Error("Failed to fetch songs with filters")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Возвращаем результат
	return c.JSON(result)
}

// GetSongText возвращает текст песни
// @Summary Получение текста песни
// @Description Возвращает текст песни с поддержкой пагинации
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path string true "ID песни"
// @Param offset query int false "Страница текста (по умолчанию 1)" default(1)
// @Param limit query int false "Количество строк текста на странице (по умолчанию 2)" default(2)
// @Success 200 {object} map[string]interface{} "Текст песни с пагинацией"
// @Failure 400 {object} map[string]interface{} "Неверный запрос (например, отсутствует ID песни)"
// @Failure 500 {object} map[string]interface{} "Ошибка сервера"
// @Router /api/songs/{id}/text [get]
func (sc *SongController) GetSongText(c *fiber.Ctx) error {
	songID := c.Params("id")
	if songID == "" {
		sc.logger.Warn("Missing song ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Song ID is required"})
	}

	// Получаем `groupID` из запроса
	groupID, err := strconv.Atoi(c.Query("group_id"))
	// if err != nil || groupID < 0 {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	// }

	// Получение параметров пагинации
	offset, err := strconv.Atoi(c.Query("offset", "1"))
	if err != nil || offset < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page number"})
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil || limit < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid limit"})
	}
	sc.logger.Infof("offset: %d, limit: %d, groupID: %d", offset, limit, groupID)

	sections, err := sc.songService.GetSongText(songID, uint(groupID), offset, limit)
	if err != nil {
		sc.logger.Errorf("Failed to fetch song text: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Преобразуем массив строк в массив объектов
	var values []map[string]string
	for _, section := range sections {
		values = append(values, map[string]string{
			"text": section,
		})
	}

	return c.JSON(fiber.Map{
		"values": values,
		"page":   offset,
		"limit":  limit,
	})
}


// DeleteSong удаляет песню
// @Summary Удаление песни
// @Description Удаляет песню по ее ID
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path string true "ID песни"
// @Success 204 "Песня успешно удалена"
// @Failure 400 {object} map[string]interface{} "Неверный запрос (например, отсутствует ID песни)"
// @Failure 500 {object} map[string]interface{} "Ошибка сервера"
// @Router /api/songs/{id} [delete]
func (sc *SongController) DeleteSong(c *fiber.Ctx) error {
	songID := c.Params("id")
	if songID == "" {
		sc.logger.Warn("Missing song ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Song ID is required"})
	}

	if err := sc.songService.DeleteSong(songID); err != nil {
		sc.logger.Error("Failed to delete song")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateSong обновляет данные песни
// @Summary Обновление данных песни
// @Description Обновляет данные песни по ее ID
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path string true "ID песни"
// @Param data body map[string]interface{} true "Данные для обновления"
// @Success 200 "Песня успешно обновлена"
// @Failure 400 {object} map[string]interface{} "Неверный запрос (например, отсутствует ID песни)"
// @Failure 500 {object} map[string]interface{} "Ошибка сервера"
// @Router /api/songs/{id} [put]
func (sc *SongController) UpdateSong(c *fiber.Ctx) error {
	songID := c.Params("id")
	if songID == "" {
		sc.logger.Warn("Missing song ID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Song ID is required"})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		sc.logger.Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := sc.songService.UpdateSong(songID, data); err != nil {
		sc.logger.Error("Failed to update song")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusOK)
}

// AddSong добавляет новую песню
// @Summary Добавление новой песни
// @Description Добавляет новую песню с указанием группы и названия
// @Tags Песни
// @Accept json
// @Produce json
// @Param data body map[string]string true "Данные для добавления песни (группа и название)"
// @Success 201 "Песня успешно добавлена"
// @Failure 400 {object} map[string]interface{} "Неверный запрос (например, отсутствует группа или название песни)"
// @Failure 500 {object} map[string]interface{} "Ошибка сервера"
// @Router /api/songs [post]
func (sc *SongController) AddSong(c *fiber.Ctx) error {
	var songData map[string]string
	if err := c.BodyParser(&songData); err != nil {
		sc.logger.Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	group := songData["group"]
	song := songData["song"]
	if group == "" || song == "" {
		sc.logger.Warn("Group or song name is missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Group and song name are required"})
	}

	if err := sc.songService.AddSong(group, song); err != nil {
		sc.logger.Error("Failed to add song")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusCreated)
}
