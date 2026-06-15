package main

import (
    "github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}

func main() {
    initDB()
    defer db.Close()

    r := gin.Default()
    r.Use(corsMiddleware())

    api := r.Group("/api")
    {
        api.GET("/user", getUser)
        api.POST("/user", updateUser)
        api.GET("/tasks", getTodayTasks)
        api.PUT("/tasks/toggle", toggleTask)
        api.POST("/checkin", manualCheckin)
        api.GET("/checkin/history", getCheckinHistory)
        api.GET("/stage-progress", getStageProgress)
        api.PUT("/stage-progress", updateStageProgress)
        api.POST("/ai-diagnose", aiDiagnose)
        api.POST("/visit", recordVisit)
        api.GET("/stats", getStats)
    }

    r.Run(SERVER_PORT)
}