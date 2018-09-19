package router

import (
	"github.com/gin-gonic/gin"
	// "net/http"
	"github.com/Zerkkro/november/src/nov/handlers"
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
	r.Run(":9100")
}
