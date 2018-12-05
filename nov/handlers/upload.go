package handlers

import (
	"html/template"
	"io"
	// "fmt"
	"github.com/gin-gonic/gin"
	"github.com/johng-cn/gf/g/os/glog"
	"os"
)

// HandleUpload ...
func HandleUpload(c *gin.Context) {
	r := c.Request
	glog.Infof("Method: %s\n", r.Method)
	if r.Method == "GET" {
		t, err := template.ParseFiles("./templates/upload.gptl")
		if err != nil {
			glog.Infof("parse template upload.gptl failed, err: %s\n", err.Error())
			return
		}
		t.Execute(c.Writer, nil)
	} else {
		file, handle, err := r.FormFile("file")
		if err != nil {
			glog.Infof("upload file failed, err: %s\n", err.Error())
			return
		}
		f, err := os.OpenFile("./upload/"+handle.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			glog.Infof("save file failed, err: %s\n", err.Error())
			return
		}
		io.Copy(f, file)
		defer f.Close()
		defer file.Close()
		glog.Infof("upload file completed.\n")
		c.Writer.Write([]byte("complete"))
	}
}
