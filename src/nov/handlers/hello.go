package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
)

// HelloPage show hello
func HelloPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "welcome",
	})
}

// tett ...
type tett struct {
	No, Title string
}

// IndexHeader ...
type indexHeader struct {
	Title string
}

func tettGet() []tett {
	t := tett{}
	ts := []tett{}

	for i := 0; i < 5; i++ {
		t.No = fmt.Sprintf("%d", i)
		t.Title = fmt.Sprintf("title%d", i)
		ts = append(ts, t)
	}
	return ts
}

// HandleIndex show index
func HandleIndex(c *gin.Context) {
	var tmpl = template.Must(template.ParseFiles("templates/index.html"))
	ts := tettGet()

	ih := &indexHeader{"Welcome"}
	err := tmpl.ExecuteTemplate(c.Writer, "Header", ih)
	if err != nil {
		fmt.Printf("failed to execute template index.html, err: %s\n", err.Error())
	}
	err = tmpl.ExecuteTemplate(c.Writer, "Index", ts)
	if err != nil {
		fmt.Printf("failed to execute template index.html, err: %s\n", err.Error())
	}
	err = tmpl.ExecuteTemplate(c.Writer, "Index", ts)
	if err != nil {
		fmt.Printf("failed to execute template index.html, err: %s\n", err.Error())
	}
}
