package main

import (
	"fmt"
	"log"
	"os"

	"xiazki/db"
	"xiazki/handler"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	h := handler.NewHandler(database)
	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		Skipper: func(c echo.Context) bool {
			return false
		},
		BeforeNextFunc: func(c echo.Context) {},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("[%v %v: %v]\n", c.Request().Method, c.Request().RequestURI, v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{TokenLookup: "header:X-CSRF-Token"}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret")))) // FIXME: SECRET!!!

	e.Static("/static", "static")

	e.GET("/login", h.GetLogin)
	e.POST("/login", h.PostLogin)
	e.GET("/register", h.GetRegister)
	e.POST("/register", h.PostRegister)

	protected := e.Group("")
	protected.Use(h.RequireAuth)
	protected.GET("/", h.GetBooks)
	protected.GET("/books", h.GetBooks)
	protected.GET("/add_book", h.GetAddBook)
	protected.GET("/book/:id", h.GetBook)

	protected_htmx := protected.Group("")
	protected_htmx.Use(h.RequireAuthHTMX)
	protected_htmx.DELETE("/book/:id", h.DeleteBook)
	protected_htmx.GET("/book/:id/edit", h.GetBookEdit)
	protected_htmx.POST("/add_book", h.PostAddBook)
	protected_htmx.POST("/logout", h.PostLogout)
	protected_htmx.PUT("/book/:id", h.PutBook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
