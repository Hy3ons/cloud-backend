package controllers

import (
	"net/http"

	"vm-controller/internal/services/k8s_service"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	K8sService *k8s_service.K8sService
}

func NewHealthController(k8sService *k8s_service.K8sService) *HealthController {
	return &HealthController{
		K8sService: k8sService,
	}
}

func (h *HealthController) Check(c *gin.Context) {
	status, err := h.K8sService.CheckConnectivity()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "ok",
		"k8s_connectivity": status,
	})
}
