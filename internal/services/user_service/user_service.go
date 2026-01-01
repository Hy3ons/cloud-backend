package userservice

import (
	"errors"
	"vm-controller/internal/db"
	"vm-controller/internal/models"
)

type UserService struct {
}

// AuthenticateUser 함수는 학번과 비밀번호를 받아 유저를 인증하고 반환합니다.
func (s *UserService) AuthenticateUser(studentID, password string) (*models.User, error) {
	database := db.GetDB()

	var user models.User
	// 학번으로 유저 찾기
	if err := database.Where("user_student_id = ?", studentID).First(&user).Error; err != nil {
		return nil, err
	}

	// 비밀번호 확인
	if !user.CheckPassword(password) {
		return nil, errors.New("비밀번호가 일치하지 않습니다")
	}

	return &user, nil
}
