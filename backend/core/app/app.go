package app

import (
	"fmt"

	"root/config"
	_ "root/core/docs"
	"root/database"
	"root/shared/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"gorm.io/gorm"
)

type App struct {
	app *fiber.App

	config     *config.Config
	logger     *logger.Logger
	httpConfig config.HTTPConfig

	db *gorm.DB

	moduleProvider *moduleProvider
}

func NewApp() *App {
	return &App{
		app: fiber.New(),
	}
}

func (app *App) Run() error {
	app.app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:6969",
		AllowCredentials: false,
	}))

	err := app.initDeps()

	if err != nil {
		return err
	}

	return app.runHttpServer()
}

func (app *App) initDeps() error {

	inits := []func() error{
		app.initConfig,
		app.initLogger,

		app.initDb,

		app.initModuleProvider,
		app.initRouter,
	}
	for _, init := range inits {
		err := init()
		if err != nil {
			return fmt.Errorf("%s", "✖ Failed to initialize dependencies: "+err.Error())
		}
	}
	return nil
}

func (app *App) initConfig() error {
	if app.config == nil {
		config, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("%s", "✖ Failed to load config: "+err.Error())
		}
		app.config = config
	}

	err := config.Load("../.env")
	if err != nil {
		return fmt.Errorf("%s", "✖ Failed to load config: "+err.Error())
	}

	return nil
}

func (app *App) initDb() error {
	if app.db == nil {
		db, err := database.ConnectDb(app.config.DatabaseUrl, app.logger)
		if err != nil {
			return err
		}
		app.db = db

		// true - запустить миграцию
		// false - не запускать
		if err := database.Migrate(db, true, app.logger); err != nil {
			return fmt.Errorf("%s", "✖ Failed to migrate database: "+err.Error())
		}
	}

	return nil
}

func (app *App) initLogger() error {
	if app.logger == nil {
		app.logger = logger.GetLogger()
	}
	return nil
}

func (app *App) initModuleProvider() error {
	err := error(nil)
	app.moduleProvider, err = NewModuleProvider(app)
	if err != nil {
		app.logger.Errorf("%s", err.Error())
		return err
	}
	return nil
}

func (app *App) runHttpServer() error {
	if app.httpConfig == nil {
		cfg, err := config.NewHTTPConfig()
		if err != nil {
			app.logger.Errorf("%s", "✖ Failed to load config: "+err.Error())
			return fmt.Errorf("✖ Failed to load config: %v", err)
		}
		app.httpConfig = cfg
	}

	app.logger.Infof("🌐 Server is running on %s", app.httpConfig.Address())
	app.logger.Info("✅ Server started successfully")
	if err := app.app.Listen(app.httpConfig.Address()); err != nil {
		app.logger.Errorf("%s", "✖ Failed to start server: "+err.Error())
		return fmt.Errorf("✖ Failed to start server: %v", err)
	}

	return nil
}

func (app *App) initRouter() error {
	api := app.app.Group("/api")

	app.app.Get("/swagger/*", fiberSwagger.WrapHandler)

	app.moduleProvider.song.InitRoutes(api)

	return nil
}
