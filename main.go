package main

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"

    "ticket-system/internal/auth"
    "ticket-system/internal/middleware"
    "ticket-system/internal/tickets"
    "ticket-system/internal/web"
)

func main() {
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    r.GET("/", func(c *gin.Context) {
        c.Data(http.StatusOK, "text/html; charset=utf-8", web.Index)
    })
    r.StaticFS("/assets", http.FS(web.Assets()))

    r.POST("/auth/register", auth.Register)
    r.POST("/auth/login", auth.Login)

    protected := r.Group("/tickets")
    protected.Use(middleware.JWTAuth())
    {
        protected.POST("", tickets.CreateTicket)
        protected.GET("", tickets.ListTickets)
        protected.GET("/:id", tickets.GetTicket)
        protected.PATCH("/:id/status", tickets.UpdateStatus)
    }

    if err := r.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
