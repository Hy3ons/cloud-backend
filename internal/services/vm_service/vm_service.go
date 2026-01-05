package vmservice

import (
	"sync"
	"vm-controller/internal/db"
	"vm-controller/internal/models"

	"errors"
	"gorm.io/gorm"
)

type VmService struct {
}

var (
	vmService *VmService
	once      sync.Once
)

func GetVmService() *VmService {
	once.Do(func() {
		vmService = &VmService{}
	})

	return vmService
}

func (vmService *VmService) FetchUserVMs(userId string) ([]models.VirtualMachine, error) {
	db := db.GetDB()

	var vms []models.VirtualMachine

	//유저 ID로 VM 찾기
	if err := db.Where("user_id = ?", userId).Find(&vms).Error; err != nil {
		return nil, err
	}

	return vms, nil
}

func (vmService *VmService) FetchVmName(vmName string) (*models.VirtualMachine, error) {
	db := db.GetDB()

	var vm models.VirtualMachine

	if err := db.Where("vm_name = ?", vmName).First(&vm).Error; err != nil {
		//하나의 행도 발견 못하면, 다음과 같은 에러를 내뱉음 Gorm
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		//그냥 에러 처리.
		return nil, err
	}

	return &vm, nil
}

func (vmService *VmService) CreateUserVM(namespace string, vmName string, vmSSHPassword string, vmHost string, vmImage string, vmPort int32) (*models.VirtualMachine, error) {
	db := db.GetDB()

	var vm models.VirtualMachine

	if err := db.Create(&vm).Error; err != nil {
		return nil, err
	}

	return &vm, nil
}

func (vmService *VmService) UpdateVmStatus(vmName string, status models.EnumVmStatus) error {
	db := db.GetDB()

	if err := db.Model(&models.VirtualMachine{}).Where("vm_name = ?", vmName).Update("status", status).Error; err != nil {
		return err
	}

	return nil
}
