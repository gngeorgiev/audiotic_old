package main

import (
	"gngeorgiev/audiotic/server/api"
	"net/http"

	"log"

	"gngeorgiev/audiotic/server/player"
	"strings"

	"os"
	"os/signal"

	"encoding/json"
	"gngeorgiev/audiotic/server/socketSessionsPool"

	"strconv"

	"gopkg.in/gin-contrib/cors.v1"
	"gopkg.in/gin-gonic/gin.v1"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

func main() {
	if err := player.Init(); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.RedirectTrailingSlash = true

	c := cors.DefaultConfig()
	c.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(c))

	m := r.Group("/meta")
	{
		m.GET("/autocomplete/*query", autocompleteHandler())
		m.GET("/search/*query", searchHandler())
	}

	p := r.Group("/player")
	{
		p.GET("/play/:provider/:id", playHandler())
		p.GET("/pause", pauseHandler())
		p.GET("/resume", resumeHandler())
		p.GET("/stop", stopHandler())
		p.GET("/status", playerStatusHandler())
		p.GET("/seek/:time", seekHandler())
		p.GET("/volume/:volume", volumeHandler())
		p.GET("/updates/*info", playerUpdatesHandler())
	}

	go func() {
		log.Fatal(r.Run(":8090"))
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	<-stopCh

	if err := player.Get().Release(); err != nil {
		log.Fatal(err)
	}
}

func autocompleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := strings.Replace(c.Param("query"), "/", "", 1)
		result, err := api.Autocomplete(query)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func searchHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := strings.Replace(c.Param("query"), "/", "", 1)
		result, err := api.Search(query)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func playHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		id := c.Param("id")
		err := api.Play(strings.ToLower(provider), id)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

func pauseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := api.Pause()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

func resumeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := api.Resume()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

func playerStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := api.Status()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

func stopHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := api.Stop()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

var playerUpdatesSessionPool = socketSessionsPool.New()

func playerUpdatesHandler() gin.HandlerFunc {
	go func() {
		pl := player.Get()
		updatedCh := make(chan struct{})
		pl.OnUpdated(updatedCh)
		for {
			select {
			case <-updatedCh:
				status, err := pl.Status()
				if err != nil {
					log.Println(err)
					continue
				}

				b, err := json.Marshal(status)
				if err != nil {
					log.Println(err)
					continue
				}

				playerUpdatesSessionPool.Send(string(b), true)
			}
		}
	}()

	return func(c *gin.Context) {
		handler := sockjs.NewHandler("/player/updates", sockjs.DefaultOptions, func(s sockjs.Session) {
			playerUpdatesSessionPool.Add(s)
		})

		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func seekHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		time := c.Param("time")
		t, err := strconv.Atoi(time)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if err := api.Seek(t); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

func volumeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		volume := c.Param("volume")
		v, err := strconv.Atoi(volume)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if err := api.Volume(v); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusOK)
	}
}