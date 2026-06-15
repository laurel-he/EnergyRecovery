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

    // 无需认证的接口
    r.POST("/api/register", register)
    r.POST("/api/login", login)
    r.POST("/api/visit", recordVisit) // 访问上报可选认证
    r.GET("/api/stats", getStats)

    // 需要认证的接口
    auth := r.Group("/api")
    auth.Use(authMiddleware())
    {
        auth.GET("/me", func(c *gin.Context) {
            c.JSON(200, gin.H{"account_id": c.GetInt("account_id"), "username": c.GetString("username")})
        })
        auth.GET("/profile", getProfile)
        auth.POST("/profile", updateProfile)
        auth.GET("/tasks", getTodayTasks)
        auth.PUT("/tasks/toggle", toggleTask)
        auth.POST("/checkin", manualCheckin)
        auth.GET("/checkin/history", getCheckinHistory)
        auth.GET("/stage-progress", getStageProgress)
        auth.PUT("/stage-progress", updateStageProgress)
        auth.POST("/ai-diagnose", aiDiagnose)
    }

    r.Run(SERVER_PORT)
}