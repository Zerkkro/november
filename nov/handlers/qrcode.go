package handlers

import (
	"github.com/gin-gonic/gin"
	// qrcode "github.com/skip2/go-qrcode"
)

// HandleQrcode ...
func HandleQrcode(c *gin.Context) {
	// c.Keys["Access-Control-Allow-Origin"] = "*"
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// var png []byte
	// png, err := qrcode.Encode("https://www.google.com", qrcode.Medium, 256)
}
