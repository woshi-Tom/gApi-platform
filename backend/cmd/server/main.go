package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/mq"
	"gapi-platform/internal/pkg/crypto"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/router"
	"gapi-platform/internal/worker"

	_ "gapi-platform/docs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	skipRabbitMQ := flag.Bool("skip-rabbitmq", false, "skip RabbitMQ connection")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := crypto.Init(cfg.Security.EncryptKey); err != nil {
		log.Fatalf("Failed to initialize crypto: %v", err)
	}
	logger.Info().Msg("Crypto initialized")

	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info().Msg("Database migrations completed")

	redisClient, err := repository.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	logger.Info().Msg("Redis connected")

	var mqClient *mq.Client

	if !*skipRabbitMQ {
		mqClient, err = mq.NewClient(&cfg.RabbitMQ)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to RabbitMQ, running without MQ")
		} else {
			defer mqClient.Close()
			logger.Info().Msg("RabbitMQ connected")

			if err := mqClient.DeclareAllQueues(); err != nil {
				logger.Warn().Err(err).Msg("Failed to declare queues")
			}

			mq.SetDefaultClient(mqClient)

			if _, err = mq.SetupDefaultConsumer(mqClient); err != nil {
				logger.Warn().Err(err).Msg("Failed to start MQ consumer")
			} else {
				logger.Info().Msg("RabbitMQ consumer started")
			}
		}
	}

	vipWorker := worker.NewVIPExpiryWorker(db.GetDB(), 1*time.Minute)
	go vipWorker.Start()
	defer vipWorker.Stop()

	r := gin.Default()

	router.SetupUserRoutes(r, cfg, db, redisClient)
	router.SetupAdminRoutes(r, cfg, db, redisClient)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info().Msgf("Starting unified API server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
