package server

import (
	"github.com/gin-gonic/gin"
	"jobocron/log"
)

func StartServer() {
	r := gin.New()
	r.Use(CustomRecovery())
	r.Use(CORSMiddleware())
	r.Use(LoggerMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	log.Info("Starting server", "port", "8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server", "error", err.Error())
	}
}
