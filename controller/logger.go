package controller

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Log(c *gin.Context) {
	traceID, _ := c.Get("trace_id")
	log.Println(traceID)
}
