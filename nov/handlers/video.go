package handlers

import (
	"fmt"
	"github.com/Zerkkro/november/nov/video"
	"github.com/gin-gonic/gin"
)

// HandleTsParse ...
func HandleTsParse(c *gin.Context) {
	var filePath = "ts/test02.ts"
	err := video.ParseTsFile(filePath, c.Writer)
	if err != nil {
		fmt.Printf("handle ts parse error: %s\n", err.Error())
	}
}
