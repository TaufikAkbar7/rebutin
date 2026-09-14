package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"rebutin/internal/delivery/http/middleware"
	"rebutin/internal/pkg/token"
	"rebutin/pkg/testutil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	csrfCookieName    = "XSRF-TOKEN"
	csrfHeaderName    = "X-XSRF-TOKEN"
	sessionCookieName = "sid"
	midError          = "next handler should not be executed when middleware aborts"
	midSuccess        = "next handler should be executed"
	expectedKey       = "secret-token-123"
)

var csrfCookie = &http.Cookie{
	Name:     csrfCookieName,
	Value:    expectedKey,
	MaxAge:   0,
	Path:     "/",
	Domain:   "",
	Secure:   true,
	HttpOnly: false,
}

var sessionCookie = &http.Cookie{
	Name:     sessionCookieName,
	Value:    expectedKey,
	MaxAge:   0,
	Path:     "/",
	Domain:   "",
	Secure:   true,
	HttpOnly: true,
}

func TestAuthMiddleware_CSRF_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 and add cookie XSRF-TOKEN", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.AddCSRFMiddleware(log))

		router.POST("/api/v1/simulation", func(ctx *gin.Context) {
			ctx.Status(http.StatusCreated)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// get cookie result from middleware
		resp := w.Result()
		var xsrfCookie *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == csrfCookieName {
				xsrfCookie = c
				break
			}
		}

		assert.NotNil(t, xsrfCookie)
		assert.NotNil(t, xsrfCookie.Value)
		assert.Equal(t, csrfCookieName, xsrfCookie.Name)
		assert.Equal(t, true, xsrfCookie.Secure)
		assert.Equal(t, false, xsrfCookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, xsrfCookie.SameSite)
		assert.Equal(t, 0, xsrfCookie.MaxAge)
		assert.Equal(t, "/", xsrfCookie.Path)
	})

	t.Run("should return 200 when method not mutation data", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.CSRFMiddleware(log))

		handlerExecuted := false
		router.GET("/api/v1/categories/:id", func(ctx *gin.Context) {
			handlerExecuted = true
			ctx.Status(http.StatusOK)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+targetID.String(), nil)
		req.AddCookie(csrfCookie)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, handlerExecuted, midSuccess)
	})

	t.Run("should return 200 when method is mutation data", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.CSRFMiddleware(log))

		router.PATCH("/api/v1/simulation/:id", func(ctx *gin.Context) {
			ctx.Status(http.StatusCreated)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/simulation/"+targetID.String(), nil)

		token, _ := token.GenerateRandomToken()
		csrf := &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			MaxAge:   0,
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: false,
		}
		req.AddCookie(csrf)
		req.Header.Set(csrfHeaderName, token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		cookie, _ := req.Cookie(csrfCookieName)
		assert.Equal(t, token, cookie.Value)
		assert.Equal(t, csrfCookieName, cookie.Name)
	})
}

func TestAuthMiddleware_CSRF_Failures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when cookie XSRF-TOKEN is missing", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.CSRFMiddleware(log))

		handlerExecuted := false
		router.GET("/api/v1/categories/:id", func(ctx *gin.Context) {
			handlerExecuted = true
			ctx.Status(http.StatusOK)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+targetID.String(), nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.False(t, handlerExecuted, midError)
		assert.Contains(t, w.Body.String(), "CSRF cookie missing")
	})

	t.Run("should return 403 when header X-XSRF-TOKEN is missing", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.CSRFMiddleware(log))

		handlerExecuted := false
		router.PATCH("/api/v1/simulation/:id", func(ctx *gin.Context) {
			handlerExecuted = true
			ctx.Status(http.StatusOK)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/simulation/"+targetID.String(), nil)
		req.AddCookie(csrfCookie)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.False(t, handlerExecuted, midError)
	})

	t.Run("should return 403 when invalid CSRF token", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.CSRFMiddleware(log))

		handlerExecuted := false
		router.PATCH("/api/v1/simulation/:id", func(ctx *gin.Context) {
			handlerExecuted = true
			ctx.Status(http.StatusOK)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/simulation/"+targetID.String(), nil)
		req.AddCookie(csrfCookie)
		req.Header.Set(csrfHeaderName, csrfCookieName)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.False(t, handlerExecuted, midError)
		assert.Contains(t, w.Body.String(), "Invalid CSRF token")
	})
}

func TestAuthMiddleware_Session_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 and add cookie sid", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.AddSessionMiddleware(log))

		router.POST("/api/v1/simulation", func(ctx *gin.Context) {
			ctx.Status(http.StatusCreated)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// get cookie result from middleware
		resp := w.Result()
		var sidCookie *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == sessionCookieName {
				sidCookie = c
				break
			}
		}

		assert.NotNil(t, sidCookie)
		assert.NotNil(t, sidCookie.Value)
		assert.Equal(t, sessionCookieName, sidCookie.Name)
		assert.Equal(t, true, sidCookie.Secure)
		assert.Equal(t, true, sidCookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, sidCookie.SameSite)
		assert.Equal(t, 0, sidCookie.MaxAge)
		assert.Equal(t, "/", sidCookie.Path)
	})

	t.Run("should return 200 when sid cookie exist", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.SessionMiddleware(log))

		handlerExecuted := false
		router.GET("/api/v1/categories/:id", func(ctx *gin.Context) {
			handlerExecuted = true
			ctx.Status(http.StatusOK)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+targetID.String(), nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, handlerExecuted, midSuccess)
	})
}

func TestAuthMiddleware_Session_Failures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when cookie sid is missing", func(t *testing.T) {
		router := gin.New()
		log, _ := testutil.SetupLogger(t)

		router.Use(middleware.SessionMiddleware(log))

		handlerExecuted := false
		router.GET("/api/v1/categories/:id", func(ctx *gin.Context) {
			handlerExecuted = true
			ctx.Status(http.StatusOK)
		})

		targetID, _ := uuid.NewV7()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+targetID.String(), nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.False(t, handlerExecuted, midError)
		assert.Contains(t, w.Body.String(), "Session cookie missing")
	})
}
