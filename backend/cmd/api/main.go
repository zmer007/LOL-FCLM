// api 是路边联盟的后端 HTTP 服务（Go + Gin）。
package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/config"
	"roadside-league/backend/internal/database"
	"roadside-league/backend/internal/handler"
	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

func main() {
	cfg := config.Load()

	// 数据库
	db, err := database.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}

	// 依赖装配
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours)
	authHandler := handler.NewAuthHandler(authService)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			response.OK(c, gin.H{"service": "lol-api", "status": "up"})
		})

		// 公开：注册 / 登录
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// 需登录
		api.GET("/me", middleware.RequiredAuth(authService), authHandler.Me)
	}

	log.Printf("api listening on %s (media: %s)", cfg.Addr, cfg.MediaRoot)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
