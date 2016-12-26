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

	"gngeorgiev/audiotic/server/history"

	"fmt"

	"gopkg.in/gin-contrib/cors.v1"
	"gopkg.in/gin-gonic/gin.v1"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

func main() {
	//var isHelp bool

	//app := cli.NewApp()
	//
	//app.Flags = []cli.Flag{
	//	cli.BoolFlag{
	//		Name:  "no-web",
	//		Usage: "Do not serve web UI",
	//	},
	//	cli.StringFlag{
	//		Name:  "web-path, w",
	//		Usage: "Specify where to server the web UI from",
	//		Value: "www",
	//	},
	//	cli.BoolFlag{
	//		Name:        "help, h",
	//		Usage:       "show help",
	//		Destination: &isHelp,
	//	},
	//}
	//
	//app.Action = func(c *cli.Context) error {
	//	initApp()
	//	return nil
	//}
	//
	//app.Run(os.Args)

	//if isHelp {
	//	return
	//}

	initApp()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	<-stopCh

	if player.Get() != nil {
		if err := player.Get().Release(); err != nil {
			log.Fatal(err)
		}
	}

	if err := history.Release(); err != nil {
		log.Fatal(err)
	}
}

func initApp() {
	if err := player.Init(); err != nil {
		log.Fatal(err)
	}

	api.Autoplay(true)

	if err := history.Init(); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.RedirectTrailingSlash = true

	c := cors.DefaultConfig()
	c.AllowAllOrigins = true
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

	h := r.Group("/history")
	{
		h.GET("/get", getHistoryHandler())
	}

	s := gin.Default()
	s.Static("/", "./www")

	go func() {
		log.Fatal(r.Run(":8090"))
	}()

	go func() {
		log.Fatal(s.Run(":8091"))
	}()
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
		err := api.Play(provider, id)
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
				fmt.Println("update")
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

func getHistoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h, err := history.Get()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, h)
	}
}
