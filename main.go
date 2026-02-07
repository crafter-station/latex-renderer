package main

import (
	"os"

	"latex-renderer/internal/handler"
	"latex-renderer/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY env var is required")
	}

	r := gin.Default()

	r.POST("/render", middleware.BearerAuth(apiKey), handler.Render)

	r.Run(":8080")
}
