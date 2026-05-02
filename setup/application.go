
			package setup

import (
	"log"
	"lead_scrapper_be/internal/config"
	"lead_scrapper_be/internal/db"
	"lead_scrapper_be/pkg/logger"

	"gorm.io/gorm"
)

type Application struct {
	DB     *gorm.DB
	Config *config.Config
	Logger logger.Logger
}

func NewApplication() *Application {

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logger.NewZapLogger(cfg)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := db.Init(cfg.DBUri)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	return &Application{
		DB:     db,
		Config: cfg,
		Logger: logger,
	}
}
