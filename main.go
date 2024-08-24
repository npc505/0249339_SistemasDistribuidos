package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	mu      sync.RWMutex
	records []Record
}

var log Log

func main() {
	r := gin.Default()
	r.POST("/", handleLogPOST)
	r.GET("/", handleLogGET)
	fmt.Println("Server running on :8080")
	r.Run(":8080")
}

func handleLogPOST(ctx *gin.Context) {
	var req struct {
		Record Record `json:"record"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.mu.Lock()
	req.Record.Offset = uint64(len(log.records))
	log.records = append(log.records, req.Record)
	log.mu.Unlock()

	fmt.Printf("Nuevo registro añadido: Offset=%d, Value=%s\n", req.Record.Offset, req.Record.Value)
	ctx.JSON(http.StatusOK, gin.H{"offset": req.Record.Offset})
}

func handleLogGET(ctx *gin.Context) {
	var req struct {
		Offset uint64 `json:"offset" form:"offset"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
			return
		}
	}

	fmt.Printf("Solicitud GET recibida para offset: %d\n", req.Offset)

	log.mu.RLock()
	defer log.mu.RUnlock()

	if req.Offset >= uint64(len(log.records)) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	record := log.records[req.Offset]
	ctx.JSON(http.StatusOK, gin.H{
		"record": gin.H{
			"value":  record.Value,
			"offset": record.Offset,
		},
	})
}
