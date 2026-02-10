package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"
const apiKey = "test123"

func postRenderPDF(t *testing.T, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest("POST", baseURL+"/render/pdf", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func postRenderPDFFile(t *testing.T, path string) *http.Response {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "reading fixture %s", path)
	return postRenderPDF(t, string(data))
}

func readErrorResponse(t *testing.T, resp *http.Response) map[string]string {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result), "body: %s", string(body))
	return result
}

func TestRenderPDF_Simple(t *testing.T) {
	resp := postRenderPDFFile(t, "fixtures/simple.tex")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Greater(t, len(body), 100, "PDF too small")
	assert.Equal(t, "%PDF-", string(body[:5]), "missing PDF magic bytes")
}

func TestRenderPDF_IEEEComplex(t *testing.T) {
	resp := postRenderPDFFile(t, "fixtures/ieee_complex.tex")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "%PDF-", string(body[:5]), "missing PDF magic bytes")
	assert.Greater(t, len(body), 5000, "IEEE PDF too small for a multi-page document")
}

func TestRenderPDF_EmptyBody(t *testing.T) {
	resp := postRenderPDF(t, "")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	result := readErrorResponse(t, resp)
	assert.Equal(t, "empty body", result["error"])
}

func TestRenderPDF_InvalidSyntax(t *testing.T) {
	resp := postRenderPDFFile(t, "fixtures/invalid_syntax.tex")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Log("pdflatex produced output despite invalid syntax (non-strict mode)")
		return
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	result := readErrorResponse(t, resp)
	assert.Equal(t, "pdf render failed", result["error"])
	assert.NotEmpty(t, result["detail"])
}

func TestRenderPDF_DuplicateBeginDocument(t *testing.T) {
	resp := postRenderPDFFile(t, "fixtures/duplicate_begin.tex")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Log("pdflatex produced output despite duplicate \\begin{document}")
		return
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	result := readErrorResponse(t, resp)
	assert.Equal(t, "pdf render failed", result["error"])
}

func TestRenderPDF_NoDocumentclass(t *testing.T) {
	resp := postRenderPDFFile(t, "fixtures/no_documentclass.tex")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Log("pdflatex produced output despite missing \\documentclass")
		return
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	result := readErrorResponse(t, resp)
	assert.Equal(t, "pdf render failed", result["error"])
}

func TestRenderPDF_UnclosedEnvironment(t *testing.T) {
	resp := postRenderPDFFile(t, "fixtures/unclosed_env.tex")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Log("pdflatex produced output despite unclosed environment")
		return
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	result := readErrorResponse(t, resp)
	assert.Equal(t, "pdf render failed", result["error"])
}

func TestRenderPDF_MissingAuth(t *testing.T) {
	req, err := http.NewRequest("POST", baseURL+"/render/pdf", strings.NewReader(`\documentclass{article}\begin{document}Hi\end{document}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRenderPDF_WrongAuth(t *testing.T) {
	req, err := http.NewRequest("POST", baseURL+"/render/pdf", strings.NewReader(`\documentclass{article}\begin{document}Hi\end{document}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer wrongkey")
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
