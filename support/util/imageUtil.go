package util

import (
	"bytes"
	"image"
	"image/png"
	logger "tools-thinker/support/logger"
)

func Image2Bytes(m image.Image) (data []byte, errinfo error) {
	b := new(bytes.Buffer)
	err := png.Encode(b, m)
	if err != nil {
		logger.Error("convert image to bytes ,error: %s", err)
		return []byte(""), err
	}
	return b.Bytes(), nil
}
