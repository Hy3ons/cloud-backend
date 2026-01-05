package main

import (
	"fmt"
	"log"

	"vm-controller/internal/api/routes"
	"vm-controller/internal/config"
	"vm-controller/internal/db"
)

func main() {
	// 1. 설정 로드 (Configuration)
	config := config.Load()

	// 2. 데이터베이스 초기화 (Database Initialization)
	db.InitDB()

	// 5. 라우터 설정 (Router)
	r := routes.SetupRouter()

	// 6. 서버 시작 (Start Server)
	log.Printf("Starting server on port %s", config.Port)
	if err := r.Run(fmt.Sprintf(":%s", config.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
