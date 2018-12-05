package router

import (
	"github.com/gin-gonic/gin"
	// "net/http"
	"github.com/Zerkkro/november/nov/handlers"
)

// Init   Start server
func Init() {
	r := gin.Default()
	// r.LoadHTMLGlob("templates/*")
	v1 := r.Group("/v1")
	{
		v1.GET("/hello", handlers.HelloPage)
		v1.GET("/index", handlers.HandleIndex)
	}
	ts := r.Group("/video")
	{
		ts.GET("/ts", handlers.HandleTsParse)
		ts.Static("/css", "video/css")
	}
	apk := r.Group("/compress")
	{
		apk.GET("/apk", handlers.HandleApkParse)
	}

	// exec := r.Group("/exec")
	// {
	// 	exec.Any("/", handlers.HandleExec)
	// }
	r.Any("/exec", handlers.HandleExec)
	r.GET("/upload", handlers.HandleUpload)
	r.POST("/upload", handlers.HandleUpload)

	r.Run(":9100")
}
