package handlers

import (
	// "fmt"
	"github.com/Zerkkro/november/src/nov/compress"
	"github.com/gin-gonic/gin"
	"github.com/johng-cn/gf/g/os/glog"
)

// HandleApkParse ...
func HandleApkParse(c *gin.Context) {
	var filePath = "./apk/01.apk"
	// c.Keys["Access-Control-Allow-Origin"] = "*"
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	err := compress.ParseApkFile(filePath, c.Writer)
	if err != nil {
		glog.Errorf("handle parse apk error: %s\n", err.Error())
	}
}
