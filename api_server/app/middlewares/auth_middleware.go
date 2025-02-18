package middlewares

import (
	"app/generated/auth"
	"app/utils"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func AuthMiddleware(f auth.StrictHandlerFunc, operationID string) auth.StrictHandlerFunc {
    return func(ctx echo.Context, i interface{}) (interface{}, error) {
        // NOTE: Cookieからtokenを取得し、JWTの復号
		tokenString, _ := ctx.Cookie("token")
		if tokenString == nil {
			return nil, echo.ErrUnauthorized
		}

		token, _ := jwt.Parse(tokenString.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_TOKEN_KEY")), nil
		})

		// NOTE: userIDをContextにセット
		var userID int
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID = int(claims["user_id"].(float64))
		}

		// NOTE: contextにuserIDを格納する
		//     : コントローラ側ではcontext.Context型のため、withValue - Valueで行う
		c := utils.NewContext(ctx.Request().Context(), userID)
		ctx.SetRequest(ctx.Request().WithContext(c))
        return f(ctx, i)
    }
}

// CSRFContextMiddleware ... CSRFトークンを context.Context に埋め込むミドルウェア
func CSRFContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// NOTE: EchoのcontextからCSRFトークンを取得
		token, ok := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve CSRF token")
		}

		// NOTE: context.Context に CSRF トークンを埋め込む
		//lint:ignore SA1029 It's ok because ContextKey
		ctx := context.WithValue(c.Request().Context(), middleware.DefaultCSRFConfig.ContextKey, token)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
