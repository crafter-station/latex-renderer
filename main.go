package main

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY env var is required")
	}

	r := gin.Default()

	r.POST("/render", func(c *gin.Context) {
		// --- Auth ---
		if err := authorize(c, apiKey); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// --- Input ---
		latex, err := c.GetRawData()
		if err != nil || len(latex) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty body"})
			return
		}

		id := uuid.NewString()
		tmpDir := os.TempDir()
		texFile := filepath.Join(tmpDir, id+".tex")
		htmlFile := filepath.Join(tmpDir, id+".html")

		if err := os.WriteFile(texFile, latex, 0600); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write temp file"})
			return
		}
		defer os.Remove(texFile)
		defer os.Remove(htmlFile)

		// --- Execute LaTeXML ---
		cmd := exec.Command(
			"latexmlc",
			texFile,
			"--dest", htmlFile,
			"--pmml",
			"--post",
			"--format=html5",
			"--whatsout=fragment",
			"--timeout=20",
		)

		cmd.Dir = tmpDir

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "latex render failed",
				"detail": stderr.String(),
			})
			return
		}

		html, err := os.ReadFile(htmlFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read output"})
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", html)
	})

	r.Run(":8080")
}

func authorize(c *gin.Context, apiKey string) error {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return errors.New("missing Authorization header")
	}

	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		return errors.New("invalid Authorization format")
	}

	if auth[len(prefix):] != apiKey {
		return errors.New("invalid API key")
	}

	return nil
}
