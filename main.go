package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Project struct {
	ProjectName string `form:"project" json:"project" binding:"required"`
	File        string `form:"file" json:"file" binding:"required"`
	Line        int    `form:"line" json:"line" binding:"required"`
}

var infos sync.Map = sync.Map{}
var pubSub sync.Map = sync.Map{}

type ResFocus struct {
	Project  Project `form:"project" json:"project" binding:"required"`
	FindFlag bool    `form:"find_flag" json:"find_flag" binding:"required"`
}

type WSConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func main() {
	r := gin.Default()
	r.GET("/focus", func(c *gin.Context) {
		project := c.Query("project")
		v, ok := infos.Load(strings.ToLower(project))
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
		err := c.ShouldBindJSON(&project)
		if err != nil {
			return
		}
		fmt.Println("name:", project.ProjectName)
		fmt.Println("path:", project.File, project.Line)
		infos.Store(strings.ToLower(project.ProjectName), project)
		c.JSON(204, gin.H{})

		go func() {
			sender, ok := pubSub.Load(strings.ToLower(project.ProjectName))
			if ok {
				if wsConn, ok := sender.(*WSConnection); ok {
					wsConn.mu.Lock()
					err := wsConn.conn.WriteJSON(ResFocus{
						Project:  project,
						FindFlag: true,
					})
					wsConn.mu.Unlock()

					if err != nil {
						fmt.Printf("failed to write message to websocket: %v\n", err)
					}
				}
			}
		}()
	})
	r.GET("/ws", func(c *gin.Context) {
		project := c.Query("project")
		upGrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed"})
			return
		}

		wsConn := &WSConnection{
			conn: ws,
			mu:   sync.Mutex{},
		}

		pubSub.Store(strings.ToLower(project), wsConn)

		defer func() {
			wsConn.mu.Lock()
			ws.Close()
			wsConn.mu.Unlock()
			pubSub.Delete(project)
		}()

		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				break
			}
		}
	})

	r.Run(":8989") // 默认监听 0.0.0.0:8080
}
