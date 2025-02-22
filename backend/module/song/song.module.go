package song_module

import (
	"root/config"
	song_controller "root/module/song/controller"
	song_repo "root/module/song/repository"
	song_service "root/module/song/service"
	"root/shared/logger"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SongModule struct {
	songController song_controller.ISongController
	songService    song_service.ISongService
	songRepository song_repo.ISongRepository
	logger         *logger.Logger
	config         *config.Config
	db             *gorm.DB
}

func NewSongModule(logger *logger.Logger, config *config.Config, db *gorm.DB) *SongModule {
	return &SongModule{
		logger: logger,
		config: config,
		db:     db,
	}
}

func (m *SongModule) SongRepository() song_repo.ISongRepository {
	if m.songRepository == nil {
		m.songRepository = song_repo.NewSongRepository(m.logger, m.db)
	}
	return m.songRepository
}

func (m *SongModule) SongController() song_controller.ISongController {
	if m.songController == nil {
		m.songController = song_controller.NewSongController(m.logger, m.SongService())
	}
	return m.songController
}

func (m *SongModule) SongService() song_service.ISongService {
	if m.songService == nil {
		m.songService = song_service.NewSongService(m.logger, m.config, m.db, m.SongRepository())
	}
	return m.songService
}

func (m *SongModule) InitRoutes(router fiber.Router) {
	song := router.Group("/song")

	//получить все
	song.Get("/", func(c *fiber.Ctx) error {
		return m.SongController().GetSongs(c)
	})

	//получить по id
	song.Get("/:id", func(c *fiber.Ctx) error {
		return m.SongController().GetSongText(c)
	})

	//создать запись(через api)
	song.Post("/", func(c *fiber.Ctx) error {
		return m.SongController().AddSong(c)
	})

	//изменить по id
	song.Put("/:id", func(c *fiber.Ctx) error {
		return m.SongController().UpdateSong(c)
	})

	//удалить по id
	song.Delete("/:id", func(c *fiber.Ctx) error {
		return m.SongController().DeleteSong(c)
	})

}
