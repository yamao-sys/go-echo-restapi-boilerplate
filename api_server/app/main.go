package main

import (
	"app/controllers"
	"app/db"
	"app/generated/auth"
	"app/generated/todos"
	"app/middlewares"
	"app/services"
	"app/utils/routers"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	loadEnv()

	dbCon := db.Init()
	e := echo.New()

	// NOTE: service層のインスタンス
	authService := services.NewAuthService(dbCon)
	todoService := services.NewTodoService(dbCon)

	// NOTE: controllerをHandlerに追加
	server := controllers.NewAuthController(authService)
	strictHandler := auth.NewStrictHandler(server, nil)

	todosServer := controllers.NewTodosController(todoService)
	
	todosMiddlewares := []todos.StrictMiddlewareFunc{middlewares.AuthMiddleware}
	todosStrictHandler := todos.NewStrictHandler(todosServer, todosMiddlewares)

	appliedMiddlewareEcho := routers.ApplyMiddlewares(e)

	auth.RegisterHandlers(appliedMiddlewareEcho, strictHandler)
	todos.RegisterHandlers(appliedMiddlewareEcho, todosStrictHandler)

	appliedMiddlewareEcho.Logger.Fatal(appliedMiddlewareEcho.Start(":" + os.Getenv("SERVER_PORT")))
}

func loadEnv() {
	envFilePath := os.Getenv("ENV_FILE_PATH")
	if envFilePath == "" {
		envFilePath = ".env"
	}
	godotenv.Load(envFilePath)
}
