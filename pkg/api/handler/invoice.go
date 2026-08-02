package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseIface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type invoiceHandler struct {
	invoiceUseCase usecaseIface.InvoiceUseCase
}

func NewInvoiceHandler(uc usecaseIface.InvoiceUseCase) handlerInterface.InvoiceHandler {
	return &invoiceHandler{invoiceUseCase: uc}
}

// GetCompanyBillingProfile godoc
//
//	@Summary		Get company billing profile (Admin)
//	@Security		BearerAuth
//	@Description	Returns the issuer details printed on every subscription invoice
//	@Tags			Admin Invoice
//	@Id				GetCompanyBillingProfile
//	@Router			/admin/company-billing-profile [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved company billing profile"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve company billing profile"
func (h *invoiceHandler) GetCompanyBillingProfile(ctx *gin.Context) {
	profile, err := h.invoiceUseCase.GetCompanyBillingProfile(ctx)
	if err != nil {
		errResponse(ctx, "Failed to retrieve company billing profile", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved company billing profile", profile)
}

// UpdateCompanyBillingProfile godoc
//
//	@Summary		Update company billing profile (Admin)
//	@Security		BearerAuth
//	@Description	Updates the issuer details printed on future subscription invoices. Already-issued invoices are unaffected.
//	@Tags			Admin Invoice
//	@Id				UpdateCompanyBillingProfile
//	@Param			input	body	request.UpdateCompanyBillingProfileRequest	true	"Billing profile"
//	@Router			/admin/company-billing-profile [put]
//	@Success		200	{object}	response.Response{}	"Successfully updated company billing profile"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to update company billing profile"
func (h *invoiceHandler) UpdateCompanyBillingProfile(ctx *gin.Context) {
	var body request.UpdateCompanyBillingProfileRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	profile := domain.CompanyBillingProfile{
		LegalName:           body.LegalName,
		GSTIN:               body.GSTIN,
		PAN:                 body.PAN,
		AddressLine1:        body.AddressLine1,
		AddressLine2:        body.AddressLine2,
		City:                body.City,
		State:               body.State,
		StateCode:           body.StateCode,
		Pincode:             body.Pincode,
		Country:             body.Country,
		SACCode:             body.SACCode,
		LogoObjectKey:       body.LogoObjectKey,
		InvoiceNumberPrefix: body.InvoiceNumberPrefix,
		FooterNote:          body.FooterNote,
		Jurisdiction:        body.Jurisdiction,
	}

	updated, err := h.invoiceUseCase.UpdateCompanyBillingProfile(ctx, profile)
	if err != nil {
		errResponse(ctx, "Failed to update company billing profile", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated company billing profile", updated)
}
