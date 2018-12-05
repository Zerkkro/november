package handlers

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"github.com/johng-cn/gf/g/os/glog"
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"
)

// CacheRule ...
type CacheRule struct {
	Line int
	Rule string
}

// CacheConfig ...
type CacheConfig struct {
	URLString    string
	MatchedRules []*CacheRule
	ErrorRules   []*CacheRule
}

// HandleExec ...
func HandleExec(c *gin.Context) {
	r := c.Request
	if r.Method == "GET" {
		t, err := template.ParseFiles("./templates/cache_config.gtpl")
		if err != nil {
			glog.Infof("parse cache_config.gptl failed, err: %s\n", err.Error())
			return
		}
		t.Execute(c.Writer, nil)
	} else {
		r.ParseForm()
		urlString := r.Form.Get("url")
		glog.Infof(">>> url: %s\n", urlString)
		if urlString == "" {
			c.Writer.Write([]byte("Empty url regexp string."))
		}

		f, err := os.Open("./upload/cache.config")
		if err != nil {
			c.Writer.Write([]byte("Cannot open cache.config"))
			return
		}
		defer f.Close()
		r := bufio.NewReader(f)
		cc := &CacheConfig{URLString: urlString}
		lineNumber := 0
		var readError error
		var line string
		for {
			if readError == io.EOF {
				break
			}
			lineNumber++
			if lineNumber > 300000 {
				break
			}
			line, readError = r.ReadString('\n')
			length := len([]rune(line))

			if readError != nil && readError != io.EOF {
				glog.Infof("------>>> err: %s, line: %s\n", readError.Error(), line)
				break
			}
			if length <= 0 || line[0] == '#' {
				continue
			}
			line = strings.Replace(line, "\t", " ", -1)
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				cr := &CacheRule{Line: lineNumber, Rule: line}
				cc.ErrorRules = append(cc.ErrorRules, cr)
				continue
			}
			index := strings.Index(parts[0], "url_regex=")
			if index < 0 {
				cr := &CacheRule{Line: lineNumber, Rule: line}
				cc.ErrorRules = append(cc.ErrorRules, cr)
				continue
			}
			urlRegexp := parts[0][index+10:]
			// glog.Infof("regexp: %s\n", urlRegexp)
			reg, err := regexp.Compile(urlRegexp)
			if err != nil {
				cr := &CacheRule{Line: lineNumber, Rule: line}
				cc.ErrorRules = append(cc.ErrorRules, cr)
				continue
			}
			reg.Longest()
			if reg.MatchString(urlString) {
				// glog.Infof("regexp matched, line: %d\n", lineNumber)
				cr := &CacheRule{Line: lineNumber, Rule: line}
				cc.MatchedRules = append(cc.MatchedRules, cr)
			}
		}
		t, err := template.ParseFiles("./templates/cache_config.gtpl")
		if err != nil {
			glog.Infof("parse cache_config.gptl failed, err: %s\n", err.Error())
			return
		}
		t.Execute(c.Writer, cc)
	}
}
