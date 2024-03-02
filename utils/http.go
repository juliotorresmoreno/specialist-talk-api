package utils

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetToken(c *gin.Context) (string, error) {
	token, _ := c.Cookie("token")
	if token == "" {
		token = c.Request.URL.Query().Get("token")
	}

	if token == "" {
		token = c.Request.Header.Get("authorization")
	}
	if len(token) > 7 && strings.ToLower(token[:6]) == "bearer" {
		token = token[7:]
	}

	if token == "" {
		return token, StatusUnauthorized
	}
	return token, nil
}

func Copy(dest gin.ResponseWriter, src io.Reader) (written int64, err error) {
	for {
		b := make([]byte, 64)
		n, err := src.Read(b)
		if n > 0 {
			written += int64(n)
			dest.Write(b[:n])
			dest.Flush()
		}

		if err == io.EOF {
			break
		}
	}
	return written, nil
}

func ParseBase64File(data string) (string, error) {
	parts := strings.Split(data, ";base64,")
	if len(parts) != 2 {
		return "", StatusBadRequest
	}
	return parts[1], nil
}
