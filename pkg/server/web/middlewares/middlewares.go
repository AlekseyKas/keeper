package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

var (
	errBadReq = echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
)

// JSONHeader -  middleware checks application header
func JSONHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get("Content-Type")
		if header != echo.MIMEApplicationJSON &&
			c.Request().RequestURI != "/" &&
			!strings.Contains(c.Request().RequestURI, "swagger") &&
			!strings.Contains(c.Request().RequestURI, "favicon") &&
			!strings.Contains(c.Request().RequestURI, "keeper.png") {
			return errBadReq
		}
		return next(c)
	}
}
