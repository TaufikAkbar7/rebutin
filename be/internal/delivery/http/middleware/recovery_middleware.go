package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rebutin/internal/delivery/http/response"
)

func RecoveryMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
					"path":  c.Request.URL.Path,
				}).Error("Panic recovered")

				response.Error(c, http.StatusInternalServerError, "Internal Server Error", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
