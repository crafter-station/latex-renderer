package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type RenderReq struct {
	Content string
	Images  map[string]ImageInput
}

type ImageInput struct {
	URL string `json:"url"`
}

func newRenderReqFromContext(c *gin.Context) (*RenderReq, error) {
	content := c.PostForm("content")
	if content == "" {
		return nil, errors.New("content is required")
	}

	images := map[string]ImageInput{}
	imagesJSON := c.PostForm("images")
	if imagesJSON != "" {
		if err := json.Unmarshal([]byte(imagesJSON), &images); err != nil {
			return nil, errors.New("invalid images json")
		}
	}

	return &RenderReq{
		Content: content,
		Images:  images,
	}, nil
}

func (req *RenderReq) downloadImages(dir string) error {
	for filename, img := range req.Images {
		resp, err := http.Get(img.URL)
		if err != nil {
			return errors.New("failed to download image: " + filename)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New("failed to download image: " + filename)
		}

		imagePath := filepath.Join(dir, filepath.Base(filename))

		file, err := os.Create(imagePath)
		if err != nil {
			return errors.New("cannot save image: " + filename)
		}

		if _, err := io.Copy(file, resp.Body); err != nil {
			file.Close()
			return errors.New("cannot save image: " + filename)
		}
		file.Close()
	}

	return nil
}
