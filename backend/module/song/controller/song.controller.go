package song_controller

import (
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

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewSongController(logger *logger.Logger, songService song_service.ISongService) ISongController {
	return &SongController{
		logger:      logger,
		songService: songService,
	}
}

// GetSongs возвращает список песен с возможностью фильтрации и пагинации
// @Summary Получение списка песен
// @Description Возвращает список песен с поддержкой фильтрации и пагинации. Фильтрация доступна по группе, названию песни, дате релиза, тексту песни и ссылке. Пагинация позволяет указать смещение и количество записей на странице.
// @Tags Песни
// @Accept json
// @Produce json
// @Param offset query int false "Смещение для пагинации. Указывает, сколько записей пропустить перед началом выборки. По умолчанию 0." default(0) minimum(0)
// @Param limit query int false "Количество записей на странице. Определяет, сколько записей вернуть в ответе. По умолчанию 10." default(10) minimum(1) maximum(100)
// @Param group query string false "Фильтр по названию группы. Возвращает песни, где название группы содержит указанную строку."
// @Param song query string false "Фильтр по названию песни. Возвращает песни, где название песни содержит указанную строку."
// @Param release_date query string false "Фильтр по дате релиза. Возвращает песни, выпущенные в указанную дату. Формат даты: YYYY-MM-DD."
// @Param text query string false "Фильтр по тексту песни. Возвращает песни, где текст песни содержит указанную строку."
// @Param link query string false "Фильтр по ссылке. Возвращает песни, где ссылка содержит указанную строку."
// @Success 200 {array} map[string]interface{} "Список песен. Каждая песня представлена в виде объекта с полями: id, group_id, group, song, release_date, text, link."
// @Failure 400 {object} map[string]interface{} "Неверный запрос. Возможные причины:
// - Некорректный формат параметра offset (должен быть целым числом >= 0).
// - Некорректный формат параметра limit (должен быть целым числом >= 1 и <= 100).
// - Некорректный формат даты release_date (должен быть в формате YYYY-MM-DD)."
// @Failure 500 {object} map[string]interface{} "Ошибка сервера. Возможные причины:
// - Внутренняя ошибка базы данных.
// - Проблемы с подключением к базе данных."
// @Router /api/song [get]
func (sc *SongController) GetSongs(c *fiber.Ctx) error {
	sc.logger.Info("GetSongs: started")
	defer sc.logger.Info("GetSongs: completed")

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		sc.logger.Warn("GetSongs: invalid offset")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid offset number",
		})
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		sc.logger.Warn("GetSongs: invalid limit")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid limit number",
		})
	}

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

	sc.logger.Infof("GetSongs: offset=%d, limit=%d, filters=%v", offset, limit, filters)

	result, err := sc.songService.GetSongs(filters, offset, limit)
	if err != nil {
		sc.logger.Errorf("GetSongs: failed to fetch songs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Message: "Failed to fetch songs",
		})
	}

	return c.JSON(Response{
		Success: true,
		Message: "Songs fetched successfully",
		Data:    result,
	})
}

// GetSongText возвращает текст песни с пагинацией
// @Summary Получение текста песни
// @Description Возвращает текст песни с поддержкой пагинации. Текст разбит на секции (куплеты, припевы), и можно указать страницу и количество строк на странице.
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path string true "ID песни. Уникальный идентификатор песни в системе."
// @Param offset query int false "Страница текста. Указывает, какую страницу текста вернуть. По умолчанию 1." default(1) minimum(1)
// @Param limit query int false "Количество строк текста на странице. Определяет, сколько строк текста вернуть. По умолчанию 1." default(1) minimum(1) maximum(10)
// @Success 200 {object} Response "Успешный ответ. Возвращает текст песни, разбитый на секции."
// @Failure 400 {object} Response "Неверный запрос. Возможные причины:
// - Отсутствует ID песни.
// - Некорректный формат параметра offset (должен быть целым числом >= 1).
// - Некорректный формат параметра limit (должен быть целым числом >= 1 и <= 10)."
// @Failure 404 {object} Response "Песня не найдена. Возможные причины:
// - Песня с указанным ID не существует."
// @Failure 500 {object} Response "Ошибка сервера. Возможные причины:
// - Внутренняя ошибка базы данных.
// - Проблемы с подключением к базе данных."
// @Router /api/song/{id} [get]
func (sc *SongController) GetSongText(c *fiber.Ctx) error {
	sc.logger.Info("GetSongText: started")
	defer sc.logger.Info("GetSongText: completed")

	songID := c.Params("id")
	if songID == "" {
		sc.logger.Warn("GetSongText: missing song ID")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Song ID is required",
		})
	}

	offset, err := strconv.Atoi(c.Query("offset", "1"))
	if err != nil || offset < 1 {
		sc.logger.Warn("GetSongText: invalid offset")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid offset",
		})
	}

	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil || limit < 1 {
		sc.logger.Warn("GetSongText: invalid limit")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid limit",
		})
	}

	sc.logger.Infof("GetSongText: songID=%s, offset=%d, limit=%d", songID, offset, limit)

	sections, err := sc.songService.GetSongText(songID, offset, limit)
	if err != nil {
		sc.logger.Errorf("GetSongText: failed to fetch song text: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Message: "Failed to fetch song text",
		})
	}

	return c.JSON(Response{
		Success: true,
		Message: "Song text fetched successfully",
		Data:    sections,
	})
}

// DeleteSong удаляет песню по её ID
// @Summary Удаление песни
// @Description Удаляет песню по её уникальному идентификатору (ID). После удаления песня больше не будет доступна в системе.
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path string true "ID песни. Уникальный идентификатор песни в системе."
// @Success 200 {object} Response "Успешный ответ. Возвращает сообщение об успешном удалении."
// @Failure 400 {object} Response "Неверный запрос. Возможные причины:
// - Отсутствует ID песни."
// @Failure 404 {object} Response "Песня не найдена. Возможные причины:
// - Песня с указанным ID не существует."
// @Failure 500 {object} Response "Ошибка сервера. Возможные причины:
// - Внутренняя ошибка базы данных.
// - Проблемы с подключением к базе данных."
// @Router /api/song/{id} [delete]
func (sc *SongController) DeleteSong(c *fiber.Ctx) error {
	sc.logger.Info("DeleteSong: started")
	defer sc.logger.Info("DeleteSong: completed")

	songID := c.Params("id")
	if songID == "" {
		sc.logger.Warn("DeleteSong: missing song ID")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Song ID is required",
		})
	}

	if err := sc.songService.DeleteSong(songID); err != nil {
		sc.logger.Errorf("DeleteSong: failed to delete song: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Message: "Failed to delete song",
		})
	}

	return c.JSON(Response{
		Success: true,
		Message: "Song deleted successfully",
	})
}

// UpdateSong обновляет данные песни
// @Summary Обновление данных песни
// @Description Обновляет данные песни по её уникальному идентификатору (ID). Можно обновить одно или несколько полей песни.
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path string true "ID песни. Уникальный идентификатор песни в системе."
// @Param data body map[string]interface{} true "Данные для обновления. Должен быть объектом JSON, содержащим поля для обновления (например, group, song, release_date, text, link)."
// @Success 200 {object} Response "Успешный ответ. Возвращает сообщение об успешном обновлении."
// @Failure 400 {object} Response "Неверный запрос. Возможные причины:
// - Отсутствует ID песни.
// - Некорректный формат данных в теле запроса."
// @Failure 404 {object} Response "Песня не найдена. Возможные причины:
// - Песня с указанным ID не существует."
// @Failure 500 {object} Response "Ошибка сервера. Возможные причины:
// - Внутренняя ошибка базы данных.
// - Проблемы с подключением к базе данных."
// @Router /api/song/{id} [put]
func (sc *SongController) UpdateSong(c *fiber.Ctx) error {
	sc.logger.Info("UpdateSong: started")
	defer sc.logger.Info("UpdateSong: completed")

	songID := c.Params("id")
	if songID == "" {
		sc.logger.Warn("UpdateSong: missing song ID")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Song ID is required",
		})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		sc.logger.Errorf("UpdateSong: failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid input",
		})
	}

	if err := sc.songService.UpdateSong(songID, data); err != nil {
		sc.logger.Errorf("UpdateSong: failed to update song: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Message: "Failed to update song",
		})
	}

	return c.JSON(Response{
		Success: true,
		Message: "Song updated successfully",
	})
}

// AddSong добавляет новую песню
// @Summary Добавление новой песни
// @Description Добавляет новую песню в систему. Для добавления необходимо указать название группы и название песни.
// @Tags Песни
// @Accept json
// @Produce json
// @Param data body map[string]string true "Данные для добавления песни. Должен быть объектом JSON, содержащим поля group и song."
// @Success 201 {object} Response "Успешный ответ. Возвращает сообщение об успешном добавлении."
// @Failure 400 {object} Response "Неверный запрос. Возможные причины:
// - Отсутствует название группы или песни.
// - Некорректный формат данных в теле запроса."
// @Failure 500 {object} Response "Ошибка сервера. Возможные причины:
// - Внутренняя ошибка базы данных.
// - Проблемы с подключением к базе данных."
// @Router /api/song [post]
func (sc *SongController) AddSong(c *fiber.Ctx) error {
	sc.logger.Info("AddSong: started")
	defer sc.logger.Info("AddSong: completed")

	var songData map[string]string
	if err := c.BodyParser(&songData); err != nil {
		sc.logger.Errorf("AddSong: failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Invalid input",
		})
	}

	group := songData["group"]
	song := songData["song"]
	if group == "" || song == "" {
		sc.logger.Warn("AddSong: group or song name is missing")
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Success: false,
			Message: "Group and song name are required",
		})
	}

	if err := sc.songService.AddSong(group, song); err != nil {
		sc.logger.Errorf("AddSong: failed to add song: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Success: false,
			Message: "Failed to add song",
		})
	}

	return c.JSON(Response{
		Success: true,
		Message: "Song added successfully",
	})
}
