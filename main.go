package main

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

type Project struct {
	ProjectName string `form:"project" json:"project" binding:"required"`
	File        string `form:"file" json:"file" binding:"required"`
}

var infos sync.Map = sync.Map{}

type ResFocus struct {
	Project  Project `form:"project" json:"project" binding:"required"`
	FindFlag bool    `form:"find_flag" json:"find_flag" binding:"required"`
}

func main() {
	r := gin.Default()
	r.GET("/focus", func(c *gin.Context) {
		project := c.Query("project")
		v, ok := infos.Load(project)
		p, o := v.(Project)
		fmt.Println(p, o)
		r := ResFocus{
			Project:  p,
			FindFlag: ok || o,
		}
		fmt.Println(r)
		c.JSON(200, r)
	})
	r.POST("/focus", func(c *gin.Context) {
		var project Project
		c.ShouldBindJSON(&project)
		fmt.Println("name:", project.ProjectName)
		fmt.Println("path:", project.File)
		infos.Store(project.ProjectName, project)
		c.JSON(204, gin.H{})

	})
	r.Run(":8989") // 默认监听 0.0.0.0:8080
}
