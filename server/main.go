package main

import (
	"gngeorgiev/audiotic/server/api"
	"net/http"

	"log"

	"gngeorgiev/audiotic/server/player"
	"strings"

	"os"
	"os/signal"

	"gopkg.in/gin-contrib/cors.v1"
	"gopkg.in/gin-gonic/gin.v1"
)

//var link = "http://d3b7.vd.aclst.com/dl.php/KMU0tzLwhbE/Developers.mp3?video_id=KMU0tzLwhbE&t=S01VMHR6THdoYkUtMTM4MjQ5ODM4Ni0xNDgxMTA3ODQ0LTg3MjUxMQ%3D%3D&exp=10-12-2016&s=8c33e323449f4c909053d1b2982c96af"

func main() {
	if err := player.Init(); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	m := r.Group("/meta")
	{
		m.GET("/autocomplete/:query", autocompleteHandler())
		m.GET("/search/:query", searchHandler())
	}

	p := r.Group("/player")
	{
		p.GET("/play/:provider/:id", playHandler())
		p.GET("/pause", pauseHandler())
		p.GET("/resume", resumeHandler())
		p.GET("/status", playerStatusHandler())
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
		query := c.Param("query")
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
		query := c.Param("query")
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
