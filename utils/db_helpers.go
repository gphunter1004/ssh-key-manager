package utils

import (
	"errors"
	"ssh-key-manager/models"
	"strings"

	"gorm.io/gorm"
)

// CheckRecordExists 레코드 존재 확인
func CheckRecordExists(model interface{}, id uint) error {
	if err := models.DB.First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("레코드를 찾을 수 없습니다")
		}
		return err
	}
	return nil
}

// CheckUserOwnership 사용자 소유권 확인
func CheckUserOwnership(tableName string, resourceID, userID uint) error {
	var count int64
	err := models.DB.Table(tableName).
		Where("id = ? AND user_id = ?", resourceID, userID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("접근 권한이 없습니다")
	}

	return nil
}

// SafeUpdate 안전한 업데이트 (변경된 필드만)
func SafeUpdate(db *gorm.DB, model interface{}, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil // 업데이트할 내용이 없음
	}

	return db.Model(model).Updates(updates).Error
}

// CheckUniqueField 필드 중복 확인
func CheckUniqueField(tableName, field string, value interface{}, excludeID ...uint) (bool, error) {
	query := models.DB.Table(tableName).Where(field+" = ?", value)

	if len(excludeID) > 0 && excludeID[0] > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	return count == 0, err
}

// BuildSearchQuery 검색 쿼리 구성
func BuildSearchQuery(db *gorm.DB, searchTerm string, fields []string) *gorm.DB {
	if searchTerm == "" || len(fields) == 0 {
		return db
	}

	searchTerm = "%" + strings.ToLower(searchTerm) + "%"

	query := db
	for i, field := range fields {
		if i == 0 {
			query = query.Where("LOWER("+field+") LIKE ?", searchTerm)
		} else {
			query = query.Or("LOWER("+field+") LIKE ?", searchTerm)
		}
	}

	return query
}

// ApplyPagination 페이징 적용
func ApplyPagination(db *gorm.DB, page, limit int) *gorm.DB {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	return db.Offset(offset).Limit(limit)
}

// ApplySort 정렬 적용
func ApplySort(db *gorm.DB, sortBy, sortOrder string, allowedFields []string) *gorm.DB {
	// 허용된 필드인지 확인
	allowed := false
	for _, field := range allowedFields {
		if field == sortBy {
			allowed = true
			break
		}
	}

	if !allowed {
		return db
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return db.Order(sortBy + " " + sortOrder)
}

// CountRecords 레코드 수 계산
func CountRecords(db *gorm.DB, model interface{}) (int64, error) {
	var count int64
	err := db.Model(model).Count(&count).Error
	return count, err
}

// TransactionWrapper 트랜잭션 래퍼
func TransactionWrapper(fn func(*gorm.DB) error) error {
	tx := models.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// BulkCreate 대량 생성
func BulkCreate(items interface{}, batchSize int) error {
	return models.DB.CreateInBatches(items, batchSize).Error
}

// SoftDeleteByID 소프트 삭제
func SoftDeleteByID(model interface{}, id uint) error {
	return models.DB.Delete(model, id).Error
}

// RestoreByID 소프트 삭제 복원
func RestoreByID(model interface{}, id uint) error {
	return models.DB.Unscoped().Model(model).Where("id = ?", id).Update("deleted_at", nil).Error
}

// HardDeleteByID 하드 삭제 (실제 삭제)
func HardDeleteByID(model interface{}, id uint) error {
	return models.DB.Unscoped().Delete(model, id).Error
}
