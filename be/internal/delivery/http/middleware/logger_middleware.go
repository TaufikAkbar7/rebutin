package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		latency := time.Since(startTime)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		entry := log.WithFields(logrus.Fields{
			"status":     statusCode,
			"latency":    latency.String(),
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"user_agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			if statusCode >= 500 {
				entry.Error("HTTP Server Error")
			} else if statusCode >= 400 {
				entry.Warn("HTTP Client Warning")
			} else {
				entry.Info("HTTP Request")
			}
		}
	}
}
