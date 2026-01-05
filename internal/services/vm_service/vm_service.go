package vmservice

import (
	"errors"
	"sync"
	"vm-controller/internal/db"
	"vm-controller/internal/models"

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

func (vmService *VmService) FetchUserVMs(userId string, containPassword bool) ([]models.VirtualMachine, error) {
	db := db.GetDB()

	var vms []models.VirtualMachine

	//유저 ID로 VM 찾기
	if err := db.Where("user_id = ?", userId).Find(&vms).Error; err != nil {
		return nil, err
	}

	if !containPassword {
		for i := range vms {
			vms[i].Password = ""
		}
	}

	return vms, nil
}

func (vmService *VmService) FetchVmName(vmName string, containPassword bool) (*models.VirtualMachine, error) {
	db := db.GetDB()

	var vm models.VirtualMachine

	if err := db.Where("name = ?", vmName).First(&vm).Error; err != nil {
		//하나의 행도 발견 못하면, 다음과 같은 에러를 내뱉음 Gorm
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		//그냥 에러 처리.
		return nil, err
	}

	if !containPassword {
		vm.Password = ""
	}

	return &vm, nil
}

func (vmService *VmService) CreateUserVM(params CreateVmParams) (*models.VirtualMachine, error) {
	db := db.GetDB()

	vm := models.VirtualMachine{
		Name:      params.VmName,
		Namespace: params.Namespace,
		Password:  params.VmPassword,
		NodePort:  params.VmSSHPort,
		DiskNum:   params.VmDiskNum,
		UserID:    params.UserID,
		Image:     params.VmImage,
		Status:    models.VmStatusProvisioning,
	}

	if err := db.Create(&vm).Error; err != nil {
		return nil, err
	}

	return &vm, nil
}

func (vmService *VmService) UpdateVmStatus(vmName string, status models.EnumVmStatus) error {
	db := db.GetDB()

	if err := db.Model(&models.VirtualMachine{}).Where("name = ?", vmName).Update("status", status).Error; err != nil {
		return err
	}

	return nil
}
