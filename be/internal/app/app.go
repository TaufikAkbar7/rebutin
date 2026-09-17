package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"rebutin/config"
	"rebutin/internal/delivery/http/handler"
	"rebutin/internal/delivery/http/middleware"
	"rebutin/internal/delivery/http/response"
	"rebutin/internal/repository/postgres"
	redisRepo "rebutin/internal/repository/redis"
	"rebutin/internal/usecase"
)

func NewHTTPHandler(
	cfg *config.Config,
	db *sqlx.DB,
	log *logrus.Logger,
	redis *redis.Client,
) *gin.Engine {
	// DI
	txManager := postgres.NewTxManager(db)
	simRepo := postgres.NewSimulationRepository(db, log)
	ticketRepo := postgres.NewTicketRepository(db, log)
	participantRepo := postgres.NewParticipantRepository(db, log)
	sessionRedisRepo := redisRepo.NewSessionCacheRepository(redis)

	simUsecase := usecase.NewSimulationUseCase(txManager, simRepo, log, ticketRepo, participantRepo, sessionRedisRepo)
	simHandler := handler.NewSimulationHandler(simUsecase)

	ticketUsecase := usecase.NewTicketUseCase(log, ticketRepo)
	ticketHandler := handler.NewTicketHandler(ticketUsecase)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// router
	router := gin.New()

	router.Use(middleware.LoggerMiddleware(log))
	router.Use(middleware.RecoveryMiddleware(log))

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Service is healthy", gin.H{
			"app_env": cfg.AppEnv,
			"status":  "UP",
		})
	})

	// API V1 Group
	v1 := router.Group("/api/v1")

	v1.POST("/simulation", middleware.AddCSRFMiddleware(log), middleware.AddSessionMiddleware(log), simHandler.Start)

	protected := v1.Group("")
	protected.Use(middleware.CSRFMiddleware(log))
	protected.Use(middleware.SessionMiddleware(log))
	protected.PATCH("/simulation/end/:id", simHandler.End)

	protected.GET("/categories/:id", ticketHandler.FindCategories)

	return router
}
