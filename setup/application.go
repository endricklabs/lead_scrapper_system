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

	// Load the configuration from the environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize the logger function
	logger, err := logger.NewZapLogger(cfg)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// Initialize the database
	db, err := db.Init(cfg.DBUri)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Queue and Worker Initializer : This is background worker process that dequeues a queue and starts the scrapping process
	queueList := queue.InitQueue(db, cfg, logger)

	// DB Outbox background poller : This is background worker process that polls the database for pending jobs and enqueues them into the active queue
	queue.StartOutboxPoller(db, queueList, cfg, logger)

	// Common Application struct for the service that contains all the necessary variables
	return &Application{
		DB:        db,
		Config:    cfg,
		Logger:    logger,
		QueueList: queueList,
	}
}
