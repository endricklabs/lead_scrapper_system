package setup

import (
	"lead_scrapper_be/internal/config"
	"lead_scrapper_be/internal/db"
	"lead_scrapper_be/pkg/logger"
	"lead_scrapper_be/pkg/queue"

	"log"

	"gorm.io/gorm"
)

type Application struct {
	DB        *gorm.DB
	Config    *config.Config
	Logger    logger.Logger
	QueueList []queue.JobQueue
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

	// Queue and Worker Initializer
	queueList := queue.InitQueue(db, cfg, logger)

	// Start the DB Outbox background poller
	queue.StartOutboxPoller(db, queueList, cfg, logger)

	return &Application{
		DB:        db,
		Config:    cfg,
		Logger:    logger,
		QueueList: queueList,
	}
}
