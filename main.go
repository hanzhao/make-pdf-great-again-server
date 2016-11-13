package main

import "github.com/labstack/echo"

func main() {
	e := echo.New()
	e.Static("/", "public")
	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err.Error())
	}
}
