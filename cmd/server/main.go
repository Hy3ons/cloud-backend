package main

import (
	"fmt"
	"log"

	"vm-controller/internal/api/controllers"
	"vm-controller/internal/api/routes"
	"vm-controller/internal/config"
	"vm-controller/internal/db"
	services "vm-controller/internal/services/k8s_service"
)

func main() {
	// 1. 설정 로드 (Configuration)
	cfg := config.Load()

	// 2. 데이터베이스 초기화 (Database Initialization)
	db.InitDB()

	// 3. 서비스 초기화 (Services)
	// K8sService는 쿠버네티스 리소스 제어를 담당합니다.
	k8sService, err := services.NewK8sService()
	if err != nil {
		log.Fatalf("Failed to initialize K8s service: %v", err)
	}

	// 4. 컨트롤러 초기화 (Controllers)
	healthController := controllers.NewHealthController(k8sService)

	// 5. 라우터 설정 (Router)
	r := routes.SetupRouter(healthController)

	// 6. 서버 시작 (Start Server)
	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
