package main

import (
	"fmt"
	"log"
	"os"

	"xiazki/internal/db"
	"xiazki/internal/handler"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if os.Getenv("APP_ENV") != "prod" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = database.Close() }()

	h := handler.NewHandler(database, os.Getenv("GOOGLE_BOOKS_API_KEY"))
	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		Skipper: func(c echo.Context) bool {
			return c.Request().RequestURI == "/favicon.ico"
		},
		BeforeNextFunc: func(c echo.Context) {},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("[%v %v: %v]\n", c.Request().Method, c.Request().RequestURI, v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSecure:   os.Getenv("APP_ENV") == "prod",
	}))
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	e.Static("/static", "web/static")
	e.File("/static/img/cover.jpeg", "assets/img/cover.jpeg")

	e.GET("/login", h.GetLogin)
	e.POST("/login", h.PostLogin)
	e.GET("/register", h.GetRegister)
	e.POST("/register", h.PostRegister)

	protected := e.Group("")
	protected.Use(h.RequireAuth)
	protected.GET("/", h.GetBooks)
	protected.GET("/books", h.GetBooks)
	protected.GET("/author/:id", h.GetAuthor)
	protected.GET("/add_book", h.GetAddBook)
	protected.GET("/book/:id", h.GetBook)
	protected.GET("/book/:id/edit", h.GetBookEdit)
	protected.GET("/profile", h.GetProfile)

	protectedHX := protected.Group("")
	protectedHX.Use(h.RequireAuthHTMX)
	protectedHX.POST("/logout", h.PostLogout)
	protectedHX.POST("/user/change_password", h.PostUserChangePassword)
	protectedHX.POST("/add_book", h.PostAddBook)
	protectedHX.GET("/add_book/autofill", h.GetAddBookAutofill)
	protected.GET("/add_book/autofill/sse", h.GetAddBookAutofillSSE)
	protectedHX.POST("/add_book/autofill/select", h.PostAddBookAutofillSelect)
	protectedHX.DELETE("/book/:id", h.DeleteBook)
	protectedHX.GET("/book/:id/stats", h.GetBookStats)
	protectedHX.PUT("/book/:id/edit", h.PutBookEdit)
	protectedHX.GET("/book/:id/add_event", h.GetBookAddEvent)
	protectedHX.POST("/book/:id/add_event", h.PostBookAddEvent)
	protectedHX.DELETE("/event/:id", h.DeleteEvent)

	e.Logger.Debug(e.Start(":8080"))
}
