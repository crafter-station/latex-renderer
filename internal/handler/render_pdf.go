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

// RenderPDF converts LaTeX source to a PDF document.
//
//	@Summary		Render LaTeX to PDF
//	@Description	Compiles a full LaTeX document into a PDF using pdflatex.
//	@Tags			render
//	@Accept			multipart/form-data
//	@Produce		application/pdf
//	@Param			Authorization	header		string	true	"Bearer API key"
//	@Param			content         formData	string	true	"LaTeX source code"
//	@Param			images          formData	string	false	"JSON map of images. Example: {\"image.jpg\":{\"url\":\"https://example.com/image.jpg\"}}"
//	@Success		200	{file}		binary	"PDF document"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/render/pdf [post]
func RenderPDF(c *gin.Context) {
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
	pdfFile := filepath.Join(tmpDir, id+".pdf")

	if err := os.WriteFile(texFile, []byte(req.Content), 0600); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "cannot write tex file"})
		return
	}

	if err := req.downloadImages(tmpDir); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "pdf render failed",
			Detail: log,
		})
		return
	}

	pdf, err := os.ReadFile(pdfFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "cannot read output"})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdf)
}

// extractTexErrors extracts LaTeX error lines starting with "!" from the compilation log.
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
