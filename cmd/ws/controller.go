package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nekomeowww/ws-repro/pkg/utils"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Type string

const (
	TypeSync    Type = "sync"
	TypeMessage Type = "message"
)

type Message struct {
	Username string `json:"username" binding:"required"`
	Message  string `json:"message" binding:"required"`
}

type Messages struct {
	Messages []Message `json:"messages" binding:"required"`
}

type WebsocketMessage[T any] struct {
	Type   Type  `json:"type"`
	RoomID int64 `json:"roomId"`
	Data   T     `json:"data"`
}

func HandleWebsocket(logger *logrus.Logger) func(*gin.Context) {
	return func(ctx *gin.Context) {
		// roomIDStr := ctx.Param("roomId") // uncomment this line to fix the issue
		w, err := utils.NewWebsocket(upgrader, ctx.Writer, ctx.Request, logger)
		if err != nil {
			logger.Error(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "invalid websocket request",
			})
			return
		}

		go func() {
			for {
				roomIDStr := ctx.Param("roomId") // comment this line to fix the issue
				roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
				if err != nil {
					logger.Error(err)
					return
				}

				err = w.Conn.WriteMessage(websocket.TextMessage, lo.Must(json.Marshal(WebsocketMessage[*Messages]{
					Type:   TypeSync,
					RoomID: roomID,
					Data: &Messages{
						Messages: []Message{{
							Username: fmt.Sprintf("Room %d", roomID),
							Message:  fmt.Sprintf("Random %d", lo.Must(rand.Int(rand.Reader, big.NewInt(1000))).Int64()),
						}},
					},
				})))
				if err != nil {
					logger.Error(err)
					w.Close()
					return
				}

				time.Sleep(time.Millisecond * 250)
			}
		}()
	}
}
