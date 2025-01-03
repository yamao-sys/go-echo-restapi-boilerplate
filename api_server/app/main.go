package main

import (
	"app/controllers"
	"app/db"
	"app/generated/auth"
	"app/generated/todos"
	"app/middlewares"
	"app/services"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	loadEnv()

	dbCon := db.Init()

	// NOTE: service層のインスタンス化
	authService := services.NewAuthService(dbCon)
	todoService := services.NewTodoService(dbCon)

	// NOTE: controllerをHandlerに追加
	server := controllers.NewAuthController(authService)
	strictHandler := auth.NewStrictHandler(server, nil)

	todosServer := controllers.NewTodosController(todoService)
	
	todosMiddlewares := []todos.StrictMiddlewareFunc{middlewares.AuthMiddleware}
	todosStrictHandler := todos.NewStrictHandler(todosServer, todosMiddlewares)

	// NOTE: Handlerをルーティングに追加
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	auth.RegisterHandlers(e, strictHandler)
	todos.RegisterHandlers(e, todosStrictHandler)

	e.Logger.Fatal(e.Start(":" + os.Getenv("SERVER_PORT")))
}

func loadEnv() {
	envFilePath := os.Getenv("ENV_FILE_PATH")
	if envFilePath == "" {
		envFilePath = ".env"
	}
	godotenv.Load(envFilePath)
}
