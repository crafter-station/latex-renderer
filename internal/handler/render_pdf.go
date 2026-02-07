package handler

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RenderPDF(c *gin.Context) {
	latex, err := c.GetRawData()
	if err != nil || len(latex) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty body"})
		return
	}

	id := uuid.NewString()
	tmpDir := os.TempDir()
	texFile := filepath.Join(tmpDir, id+".tex")
	pdfFile := filepath.Join(tmpDir, id+".pdf")

	if err := os.WriteFile(texFile, latex, 0600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write temp file"})
		return
	}
	defer os.Remove(texFile)
	defer os.Remove(pdfFile)
	defer os.Remove(filepath.Join(tmpDir, id+".aux"))
	defer os.Remove(filepath.Join(tmpDir, id+".log"))

	cmd := exec.Command(
		"pdflatex",
		"-interaction=nonstopmode",
		"-output-directory", tmpDir,
		"-jobname", id,
		texFile,
	)

	cmd.Dir = tmpDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log := stderr.String()
		if log == "" {
			if logBytes, e := os.ReadFile(filepath.Join(tmpDir, id+".log")); e == nil {
				log = extractTexErrors(string(logBytes))
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "pdf render failed",
			"detail": log,
		})
		return
	}

	pdf, err := os.ReadFile(pdfFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read output"})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdf)
}

func extractTexErrors(log string) string {
	var errors []string
	for _, line := range strings.Split(log, "\n") {
		if strings.HasPrefix(line, "!") {
			errors = append(errors, line)
		}
	}
	if len(errors) == 0 {
		return log
	}
	return strings.Join(errors, "\n")
}
