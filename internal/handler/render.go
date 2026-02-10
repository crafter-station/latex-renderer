package handler

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//go:embed static/css/LaTeXML.css
var latexmlCSS string

// Render converts LaTeX source to HTML with embedded CSS.
//
//	@Summary		Render LaTeX to HTML
//	@Description	Converts a full LaTeX document into an HTML fragment with embedded LaTeXML CSS and Presentation MathML.
//	@Tags			render
//	@Accept			text/plain
//	@Produce		text/html
//	@Param			Authorization	header	string	true	"Bearer API key"
//	@Param			body			body	string	true	"LaTeX source code"
//	@Success		200	{string}	string	"HTML with embedded CSS"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/render [post]
func Render(c *gin.Context) {
	latex, err := c.GetRawData()
	if err != nil || len(latex) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "empty body"})
		return
	}

	id := uuid.NewString()
	tmpDir := os.TempDir()
	texFile := filepath.Join(tmpDir, id+".tex")
	htmlFile := filepath.Join(tmpDir, id+".html")

	if err := os.WriteFile(texFile, latex, 0600); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "cannot write temp file"})
		return
	}
	defer os.Remove(texFile)
	defer os.Remove(htmlFile)

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
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "latex render failed",
			Detail: stderr.String(),
		})
		return
	}

	html, err := os.ReadFile(htmlFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "cannot read output"})
		return
	}

	styled := fmt.Sprintf("<style>\n%s\n</style>\n%s", latexmlCSS, html)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(styled))
}
