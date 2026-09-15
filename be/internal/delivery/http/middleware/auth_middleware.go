package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rebutin/internal/delivery/http/response"
	"rebutin/internal/pkg/token"
)

const (
	csrfCookieName    = "XSRF-TOKEN"
	csrfHeaderName    = "X-XSRF-TOKEN"
	sessionCookieName = "sid"
)

func CSRFMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, _ := c.Cookie(csrfCookieName)
		if cookie == "" {
			log.Errorf("[Middelware.CSRFMiddleware] CSRF is missing")
			response.Error(c, http.StatusForbidden, "CSRF cookie missing", nil)
			c.Abort()
			return
		}

		// bypass if not method mutation data
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			c.Next()
			return
		}

		// get csrf
		csrf := c.GetHeader(csrfHeaderName)
		if csrf == "" {
			response.Error(c, http.StatusForbidden, "CSRF token missing in header", nil)
			c.Abort()
			return
		}

		// compare cookie
		if subtle.ConstantTimeCompare([]byte(cookie), []byte(csrf)) != 1 {
			response.Error(c, http.StatusForbidden, "Invalid CSRF token", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

func SessionMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, _ := c.Cookie(sessionCookieName)
		if cookie == "" {
			log.Errorf("[Middelware.SessionMiddleware] Cookie is missing")
			response.Error(c, http.StatusForbidden, "Session cookie missing", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

func AddSessionMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, _ := c.Cookie(sessionCookieName)
		if cookie == "" {
			token, err := token.GenerateRandomToken()
			if err != nil {
				log.Errorf("[Middelware.AddSessionMiddleware] Error get session: %v", err)
				response.Error(c, http.StatusInternalServerError, "Internal Server Error", nil)
				c.Abort()
				return
			}

			// set cookie
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(sessionCookieName, token, 0, "/", "", true, true)
			c.Set("session-cookie", token)
			log.Infof("[Middelware.AddSessionMiddleware] Set session cookie successfully")
		}

		c.Next()
	}
}

func AddCSRFMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, _ := c.Cookie(csrfCookieName)
		if cookie == "" {
			token, err := token.GenerateRandomToken()
			if err != nil {
				log.Errorf("[Middelware.AddCSRFMiddleware] Error get CSRF: %v", err)
				response.Error(c, http.StatusInternalServerError, "Internal Server Error", nil)
				c.Abort()
				return
			}

			// set cookie
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(csrfCookieName, token, 0, "/", "", true, false)
			log.Infof("[Middelware.AddCSRFMiddleware] Set cookie CSRF successfully")
		}

		c.Next()
	}
}
