package main

import (
	"apigo/runner/models"
	"apigo/runner/services"
	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"
	"net/http"
)

func main() {
	app := gin.Default()
	app.Use(favicon.New("./favicon.ico"))
	app.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello favicon.")
	})
	app.GET("/run", func(c *gin.Context) {
		config := models.NewRunner(100, 100)
		services.RunnerRun(*config)
		c.String(http.StatusOK, "Runner")
	})
	app.Run(":8080")
}
