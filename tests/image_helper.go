package tests

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/image/draw"
)

type ImageOptimizer struct {
	MaxDimension int
	Quality      int
}

func NewImageOptimizer() *ImageOptimizer {
	return &ImageOptimizer{
		MaxDimension: 2048,
		Quality:      85,
	}
}

func (io *ImageOptimizer) OptimizeAndEncode(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	resized := io.resize(img)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: io.Quality}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (io *ImageOptimizer) resize(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= io.MaxDimension && height <= io.MaxDimension {
		return img
	}

	var newWidth, newHeight int
	if width > height {
		newWidth = io.MaxDimension
		newHeight = int(float64(height) * float64(io.MaxDimension) / float64(width))
	} else {
		newHeight = io.MaxDimension
		newWidth = int(float64(width) * float64(io.MaxDimension) / float64(height))
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}
