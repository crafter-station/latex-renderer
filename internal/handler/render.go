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
//	@Accept			multipart/form-data
//	@Produce		text/html
//	@Param			Authorization	header		string	true	"Bearer API key"
//	@Param			content         formData	string	true	"LaTeX source code"
//	@Param			images          formData	string	false	"JSON map of images. Example: {\"image.jpg\":{\"url\":\"https://example.com/image.jpg\"}}"
//	@Success		200	{string}	string	"HTML with embedded CSS"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/render [post]
func Render(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid form"})
		return
	}

	req, err := newRenderReqFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	id := uuid.NewString()
	tmpDir := os.TempDir()
	texFile := filepath.Join(tmpDir, id+".tex")
	htmlFile := filepath.Join(tmpDir, id+".html")

	if err := os.WriteFile(texFile, []byte(req.Content), 0600); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "cannot write temp file"})
		return
	}

	if err := req.downloadImages(tmpDir); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	defer os.Remove(texFile)
	defer os.Remove(htmlFile)
	defer os.Remove(filepath.Join(tmpDir, id+".aux"))
	defer os.Remove(filepath.Join(tmpDir, id+".log"))

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
		log := stderr.String()
		if log == "" {
			if logBytes, e := os.ReadFile(filepath.Join(tmpDir, id+".log")); e == nil {
				log = extractTexErrors(string(logBytes))
			}
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "render failed",
			Detail: log,
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
