package main

import (
	"app/controllers"
	"app/db"
	"app/generated/auth"
	"app/services"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	dbCon := db.Init()

	// NOTE: service層のインスタンス化
	authService := services.NewAuthService(dbCon)

	// NOTE: controllerをHandlerに追加
	server := controllers.NewAuthController(authService)
	strictHandler := auth.NewStrictHandler(server, nil)

	// NOTE: Handlerをルーティングに追加
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	auth.RegisterHandlers(e, strictHandler)

	e.Logger.Fatal(e.Start(":" + os.Getenv("SERVER_PORT")))
}
