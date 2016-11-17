package main

import (
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const chunkSize = 8192

func uploadHandler(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// Source
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	filePath := "/pdf/" + c.Param("name")

	// Destination
	dst, err := os.Create("public" + filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"ok":  true,
		"url": filePath,
	})
}

func validHandler(c echo.Context) error {
	filePath := "/pdf/" + c.Param("name")

	// Exist
	if _, err := os.Stat("public" + filePath); err == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":  true,
			"url": filePath,
		})
	}

	// Not exist
	return c.JSON(http.StatusNotFound, map[string]interface{}{
		"ok": false,
	})
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/upload/:name", uploadHandler)
	e.GET("/valid/:name", validHandler)
	e.Static("/", "public")
	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err.Error())
	}
}
