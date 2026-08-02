package interfaces

import "github.com/gin-gonic/gin"

type InvoiceHandler interface {
	GetCompanyBillingProfile(ctx *gin.Context)
	UpdateCompanyBillingProfile(ctx *gin.Context)
}
