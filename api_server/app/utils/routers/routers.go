package routers

import (
	"app/middlewares"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func ApplyMiddlewares(e *echo.Echo) *echo.Echo {
	// NOTE: CORSの設定
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept,echo.HeaderAccessControlAllowHeaders,echo.HeaderXCSRFToken},
	}))

	// NOTE: CSRF対策
	csrfConfig := middleware.CSRFConfig{
		TokenLookup: "header:"+echo.HeaderXCSRFToken,
		CookieMaxAge: 3600,
	}
	e.Use(middleware.CSRFWithConfig(csrfConfig))

	// NOTE: CSRF トークンを context.Context に埋め込むミドルウェアを適用
	// 	   : StrictHandlerだとecho.Contextがhandler側で使えずのため
	e.Use(middlewares.CSRFContextMiddleware)

	return e
}
