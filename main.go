package main

import (
	"os"

	_ "latex-renderer/docs"
	"latex-renderer/internal/handler"
	"latex-renderer/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			LaTeX Renderer API
//	@version		1.0
//	@description	API for converting LaTeX documents to HTML and PDF.

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer token (e.g. "Bearer your-api-key")

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY env var is required")
	}

	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/render", middleware.BearerAuth(apiKey), handler.Render)
	r.POST("/render/pdf", middleware.BearerAuth(apiKey), handler.RenderPDF)

	r.Run(":8080")
}
