package service

import (
	"errors"
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"
	"strings"

	"gorm.io/gorm"
)

// UpdateUserProfile은 사용자 프로필을 업데이트합니다.
func UpdateUserProfile(userID uint, req model.UserUpdateRequest) (*model.User, error) {
	log.Printf("✏️ 사용자 프로필 업데이트 (ID: %d)", userID)

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		return nil, err
