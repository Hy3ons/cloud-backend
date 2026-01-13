package controllers

import (
	http "net/http"
	sync "sync"
	"vm-controller/internal/middleware"
	k8s_service "vm-controller/internal/services/k8s_service"
	userservice "vm-controller/internal/services/user_service"
	vm_service "vm-controller/internal/services/vm_service"

	gin "github.com/gin-gonic/gin"
	cast "github.com/spf13/cast"
)

type VirtualMachineController struct {
	k8sService  *k8s_service.K8sService
	userService *userservice.UserService
	vmService   *vm_service.VmService
}

var (
	virtualMachineController *VirtualMachineController
	onceVM                   sync.Once
)

func (vmC *VirtualMachineController) RegisterRoutes(r *gin.RouterGroup) {
	vm := r.Group("/vm", middleware.AuthGuard())

	vm.POST("/create", vmC.CreateVM)
	vm.GET("/fetch", vmC.FetchUserVMs)
	vm.POST("/stop", vmC.StopVM)
	vm.DELETE("/delete", vmC.DeleteVM)
	vm.POST("/start", vmC.StartVM)
}

func GetVirtualMachineController() *VirtualMachineController {
	//Sync Once 는 다른 고루틴이 once.Do를 빠져 나오지 못하도록 설계되어 있음.
	onceVM.Do(func() {
		k8s_service, err := k8s_service.GetK8sService()

		if err != nil {
			// Injection 에러가 일어남.
			panic(err)
		}

		virtualMachineController = &VirtualMachineController{
			k8sService: k8s_service,
		}
	})

	// 무조건 할당되어 있습니다.
	return virtualMachineController
}

type CreateVMParams struct {
	VmName        string `json:"vm_name"`
	VmSSHPassword string `json:"vm_ssh_password"`
	VmImage       string `json:"vm_image"`
	VmHost        string `json:"vm_host"`
}

func (vmC *VirtualMachineController) CreateVM(c *gin.Context) {
	var req CreateVMParams

	user_id, _ := c.Get("user_id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, _ := vmC.userService.FetchUserById(user_id.(string))

	var signed_port int32

	vm, err := vmC.k8sService.CreateUserVM(user.Namespace,
		req.VmName, req.VmSSHPassword, req.VmHost, "yaml-data/client-vm", signed_port)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create VM"})
		return
	}

	//database 등록 절차를 가져야함.

	c.JSON(http.StatusOK, gin.H{"vm": vm})
}

func (vmC *VirtualMachineController) FetchUserVMs(c *gin.Context) {
	user_id, ok := c.Get("user_id")

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	vms, err := vmC.vmService.FetchUserVMs(user_id.(string), false)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch VMs"})
		return
	}

	// Password Is Not Sent To Client
	c.JSON(http.StatusOK, gin.H{"vms": vms})
}

type StopVMParams struct {
	VmName string `json:"vm_name"`
}

func (vmC *VirtualMachineController) StopVM(c *gin.Context) {
	user_id, _ := c.Get("user_id")

	var req StopVMParams

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	u64, err := cast.ToUintE(user_id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id controller 121:30 Switching Error"})
		return
	}

	vm, _ := vmC.vmService.FetchVmName(req.VmName, false)
	if vm.UserID != u64 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	go vmC.k8sService.StopVM(vm)

	c.JSON(http.StatusOK, gin.H{"vm": vm})
}

type StartVMParams struct {
	VmName string `json:"vm_name"`
}

func (vmC *VirtualMachineController) StartVM(c *gin.Context) {
	user_id, _ := c.Get("user_id")

	var req StartVMParams

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	u64, err := cast.ToUintE(user_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id controller 155:30 Switching Error"})
		return
	}

	vm, _ := vmC.vmService.FetchVmName(req.VmName, false)
	if vm.UserID != u64 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	go vmC.k8sService.StartVM(vm)

	c.JSON(http.StatusOK, gin.H{"vm": vm})
}

type DeleteVMParams struct {
	VmName string `json:"vm_name"`
}

func (vmC *VirtualMachineController) DeleteVM(c *gin.Context) {
	user_id, _ := c.Get("user_id")

	var req DeleteVMParams

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	u64, err := cast.ToUintE(user_id)

	// 파싱 오류 확인.
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id controller 155:30 Switching Error"})
		return
	}

	vm, _ := vmC.vmService.FetchVmName(req.VmName, false)
	// 소유권 확인.
	if vm.UserID != u64 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	go vmC.k8sService.DeleteVM(vm)

	c.JSON(http.StatusOK, gin.H{"vm": vm})
}
