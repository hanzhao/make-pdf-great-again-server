package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
	"gopkg.in/redis.v5"
)

const chunkSize = 8192

type (
	Highlight struct {
		Page  int      `json:"page"`
		Begin Position `json:"begin"`
		End   Position `json:"end"`
	}
	Position struct {
		DivIdx int `json:"divIdx"`
		Offset int `json:"offset"`
	}
	Note struct {
		Page    int    `json:"page"`
		Content string `json:"content"`
		Time    int64  `json:"time"`
	}
)

var conns map[string][]*websocket.Conn = map[string][]*websocket.Conn{}

var redisClient *redis.Client

// Upload a pdf file to server.
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

// Check if file is available in server.
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

func wsHandler(c echo.Context) error {
	name := c.Param("name")

	key := "pdf:" + name + ":highlight"

	if conns[name] == nil {
		conns[name] = make([]*websocket.Conn, 0)
	}

	websocket.Handler(func(ws *websocket.Conn) {
		conns[name] = append(conns[name], ws)
		for {
			// Send all highlights
			highlights, err := redisClient.Get(key).Result()
			if err == redis.Nil {
				highlights = "[]"
			} else if err != nil {
				log.Fatal(err)
			}
			err = websocket.Message.Send(ws, highlights)
			// Hold connection
			if err != nil {
				break
			}
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				break
			}
		}
		for i := 0; i < len(conns[name]); i++ {
			if conns[name][i] == ws {
				conns[name][i] = conns[name][len(conns[name])-1]
				conns[name] = conns[name][:len(conns[name])-1]
				break
			}
		}
		ws.Close()
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

func highlightHandler(c echo.Context) error {
	name := c.Param("name")

	key := "pdf:" + name + ":highlight"

	if conns[name] == nil {
		conns[name] = make([]*websocket.Conn, 0)
	}

	highlight := new(Highlight)
	value, err := redisClient.Get(key).Bytes()
	if err == redis.Nil {
		value = []byte{0x91, 0x93}
	} else if err != nil {
		log.Fatal(err)
	}
	var highlights []*Highlight
	json.Unmarshal(value, &highlights)

	if err := c.Bind(highlight); err == nil {
		// Save to redis
		highlights = append(highlights, highlight)
		value, _ = json.Marshal(highlights)
		redisClient.Set(key, value, 0)

		c.JSON(http.StatusOK, map[string]bool{"ok": true})

		newHighlights := make([]*Highlight, 1)
		newHighlights[0] = highlight
		bytes, _ := json.Marshal(newHighlights)
		jsonStr := string(bytes)
		// Notify to all connection
		for i := 0; i < len(conns[name]); i++ {
			websocket.Message.Send(conns[name][i], jsonStr)
		}
		return nil
	} else {
		c.JSON(http.StatusOK, map[string]bool{"ok": false})
		return err
	}
}

func addNoteHandler(c echo.Context) error {
	name := c.Param("name")

	key := "pdf:" + name + ":note"
	page, _ := strconv.Atoi(c.QueryParam("page"))

	content := c.FormValue("content")
	value, err := redisClient.Get(key).Bytes()
	if err == redis.Nil {
		value = []byte{0x91, 0x93}
	} else if err != nil {
		log.Fatal(err)
	}
	var notes []*Note
	json.Unmarshal(value, &notes)

	// Save to redis
	notes = append(notes, &Note{
		Page:    page,
		Content: content,
		Time:    time.Now().Unix(),
	})
	value, _ = json.Marshal(notes)
	redisClient.Set(key, value, 0)

	c.JSON(http.StatusOK, map[string]bool{"ok": true})
	return nil
}

func getNotesHandler(c echo.Context) error {
	name := c.Param("name")

	key := "pdf:" + name + ":note"
	value, err := redisClient.Get(key).Result()
	if err == redis.Nil {
		value = "[]"
	} else if err != nil {
		log.Fatal(err)
	}

	c.String(http.StatusOK, value)
	return nil
}

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "public")
	e.POST("/upload/:name", uploadHandler)
	e.GET("/valid/:name", validHandler)
	e.GET("/watch/:name", wsHandler)
	e.POST("/highlight/:name", highlightHandler)
	e.POST("/note/:name", addNoteHandler)
	e.GET("/note/:name", getNotesHandler)

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err.Error())
	}
}
