package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"vm-controller/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB   *gorm.DB
	once sync.Once
)

// InitDB는 환경 변수를 사용하여 데이터베이스 연결을 초기화합니다.
// 싱글톤 패턴을 사용하여 연결 풀이 하나만 생성되도록 보장합니다.
func InitDB() {
	once.Do(func() {
		// 환경 변수에서 DB 접속 정보 로드
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Seoul",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)

		var err error
		// 1. GORM을 사용하여 PostgreSQL 드라이버로 연결
		// 필요시 재시도 로직을 추가하여 안정성을 높일 수 있음
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatalf("Failed to connect to database (DB 연결 실패): %v", err)
		}

		// 2. Connection Pool(커넥션 풀) 설정
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatalf("Failed to get generic database object: %v", err)
		}

		// SetMaxIdleConns: 유휴 상태로 유지할 최대 커넥션 수
		sqlDB.SetMaxIdleConns(10)
		// SetMaxOpenConns: 데이터베이스에 오픈할 수 있는 최대 커넥션 수
		sqlDB.SetMaxOpenConns(100)
		// SetConnMaxLifetime: 커넥션이 재사용될 수 있는 최대 시간
		sqlDB.SetConnMaxLifetime(time.Hour)

		log.Println("Successfully connected to PostgreSQL database (PostgreSQL 연결 성공)")

		// 3. Auto Migration (자동 마이그레이션)
		// 정의된 모델(struct)을 기반으로 테이블을 자동으로 생성하거나 스키마를 업데이트합니다.
		log.Println("Running AutoMigrate... (테이블 자동 생성 중)")
		err = DB.AutoMigrate(
			&models.User{},
			&models.VirtualMachine{},
			&models.Deployment{},
		)
		if err != nil {
			log.Fatalf("Failed to migrate database schema: %v", err)
		}
		log.Println("Database migration completed (마이그레이션 완료)")
	})
}

// GetDB는 싱글톤 데이터베이스 인스턴스를 반환합니다.
func GetDB() *gorm.DB {
	if DB == nil {
		InitDB()
	}
	return DB
}
