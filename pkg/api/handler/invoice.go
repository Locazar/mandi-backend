package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseIface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
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

// ListMyInvoices godoc
//
//	@Summary		List my invoices (User)
//	@Security		BearerAuth
//	@Description	Returns the authenticated seller's tax invoices, most recent first
//	@Tags			User Subscription
//	@Id				ListMyInvoices
//	@Router			/subscriptions/invoices [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved invoices"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve invoices"
func (h *invoiceHandler) ListMyInvoices(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	pagination := request.GetPagination(ctx)

	invoices, err := h.invoiceUseCase.ListInvoicesForUser(ctx, userID, pagination)
	if err != nil {
		errResponse(ctx, "Failed to retrieve invoices", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved invoices", invoices)
}

// DownloadInvoice godoc
//
//	@Summary		Download an invoice (User)
//	@Security		BearerAuth
//	@Description	Returns a short-lived presigned URL for the invoice PDF
//	@Tags			User Subscription
//	@Id				DownloadInvoice
//	@Param			invoice_id	path	string	true	"Invoice ID"
//	@Router			/subscriptions/invoices/{invoice_id}/download [get]
//	@Success		200	{object}	response.Response{}	"Successfully generated download link"
//	@Failure		403	{object}	response.Response{}	"Invoice belongs to another user"
//	@Failure		404	{object}	response.Response{}	"Invoice not found"
func (h *invoiceHandler) DownloadInvoice(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	invoiceID := ctx.Param("invoice_id")

	result, err := h.invoiceUseCase.GetInvoiceDownload(ctx, invoiceID, userID, false)
	if err != nil {
		errResponse(ctx, "Failed to generate invoice download link", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully generated download link", result)
}

// ListInvoices godoc
//
//	@Summary		List all invoices (Admin)
//	@Security		BearerAuth
//	@Description	Lists subscription tax invoices with optional user and date-range filters
//	@Tags			Admin Invoice
//	@Id				ListInvoices
//	@Param			user_id	query	string	false	"Filter by subscriber id"
//	@Param			from	query	string	false	"Start date, RFC3339 or YYYY-MM-DD"
//	@Param			to		query	string	false	"End date, RFC3339 or YYYY-MM-DD"
//	@Router			/admin/subscription-invoices [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved invoices"
//	@Failure		400	{object}	response.Response{}	"Invalid date filter"
func (h *invoiceHandler) ListInvoices(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	filter := domain.InvoiceFilter{
		UserID: ctx.Query("user_id"),
		Limit:  int(pagination.Limit),
		Offset: int(pagination.Offset),
	}

	if raw := ctx.Query("from"); raw != "" {
		t, err := parseFilterDate(raw)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid 'from' date", err, nil)
			return
		}
		filter.From = &t
	}
	if raw := ctx.Query("to"); raw != "" {
		t, err := parseFilterDate(raw)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid 'to' date", err, nil)
			return
		}
		// Make the range inclusive of the whole end day.
		endOfDay := t.Add(24*time.Hour - time.Nanosecond)
		filter.To = &endOfDay
	}

	invoices, err := h.invoiceUseCase.ListInvoicesForAdmin(ctx, filter)
	if err != nil {
		errResponse(ctx, "Failed to retrieve invoices", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved invoices", invoices)
}

// AdminDownloadInvoice godoc
//
//	@Summary		Download any invoice (Admin)
//	@Security		BearerAuth
//	@Description	Returns a short-lived presigned URL for any seller's invoice PDF
//	@Tags			Admin Invoice
//	@Id				AdminDownloadInvoice
//	@Param			invoice_id	path	string	true	"Invoice ID"
//	@Router			/admin/subscription-invoices/{invoice_id}/download [get]
//	@Success		200	{object}	response.Response{}	"Successfully generated download link"
//	@Failure		404	{object}	response.Response{}	"Invoice not found"
func (h *invoiceHandler) AdminDownloadInvoice(ctx *gin.Context) {
	invoiceID := ctx.Param("invoice_id")

	result, err := h.invoiceUseCase.GetInvoiceDownload(ctx, invoiceID, "", true)
	if err != nil {
		errResponse(ctx, "Failed to generate invoice download link", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully generated download link", result)
}

// parseFilterDate accepts both a plain date and a full RFC3339 timestamp, so
// the admin UI can send either.
func parseFilterDate(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}
