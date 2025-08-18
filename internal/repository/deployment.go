package repository

import (
	"ssh-key-manager/internal/model"
)

// DeploymentRepositoryImpl Deployment Repository 구현체
type DeploymentRepositoryImpl struct {
	*BaseRepository
}

// NewDeploymentRepository Deployment Repository 생성자
func NewDeploymentRepository() (DeploymentRepository, error) {
	base, err := NewBaseRepository()
	if err != nil {
		return nil, err
	}
	return &DeploymentRepositoryImpl{BaseRepository: base}, nil
}

// Create 배포 기록 생성
func (dr *DeploymentRepositoryImpl) Create(deployment *model.ServerKeyDeployment) error {
	return dr.db.Create(deployment).Error
}

// FindByUserID 사용자 ID로 배포 기록 조회
func (dr *DeploymentRepositoryImpl) FindByUserID(userID uint) ([]model.ServerKeyDeployment, error) {
	var deployments []model.ServerKeyDeployment
	err := dr.db.Where("user_id = ?", userID).
		Preload("Server").
		Preload("SSHKey").
		Order("created_at DESC").
		Find(&deployments).Error
	return deployments, err
}

// DeleteBySSHKeyID SSH 키 ID로 배포 기록 삭제
func (dr *DeploymentRepositoryImpl) DeleteBySSHKeyID(keyID uint) error {
	return dr.db.Where("ssh_key_id = ?", keyID).Delete(&model.ServerKeyDeployment{}).Error
}

// DeleteByServerID 서버 ID로 배포 기록 삭제
func (dr *DeploymentRepositoryImpl) DeleteByServerID(serverID uint) error {
	return dr.db.Where("server_id = ?", serverID).Delete(&model.ServerKeyDeployment{}).Error
}
