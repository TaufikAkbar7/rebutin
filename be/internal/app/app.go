package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"rebutin/config"
	"rebutin/internal/delivery/http/handler"
	"rebutin/internal/delivery/http/middleware"
	"rebutin/internal/delivery/http/response"
	"rebutin/internal/repository/postgres"
	"rebutin/internal/usecase"
	customValidator "rebutin/pkg/validator"
)

func NewHTTPHandler(
	cfg *config.Config,
	db *sqlx.DB,
	log *logrus.Logger,
	val *customValidator.CustomValidator,
) *gin.Engine {
	// DI
	txManager := postgres.NewTxManager(db)
	simRepo := postgres.NewSimulationRepository(db, log)
	simUsecase := usecase.NewSimulationUseCase(txManager, simRepo, log)
	simHandler := handler.NewSimulationHandler(*simUsecase, val)

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
	v1.Use(middleware.CSRFMiddleware(log))
	v1.Use(middleware.SessionMiddleware(log))

	simulation := v1.Group("/simulation")
	simulation.Use(middleware.AddCSRFMiddleware(log))
	simulation.Use(middleware.AddSessionMiddleware(log))
	simulation.POST("", simHandler.Start)
	simulation.PATCH("/end/:id", simHandler.End)

	return router
}
