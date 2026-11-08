package request

type SubmitLeadRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Requirement string `json:"requirement" binding:"required"`
	City        string `json:"city" binding:"required"`
	Source      string `json:"source"`
}

type UpdateLeadStatusRequest struct {
	Status    string `json:"status"`
	AdminNote string `json:"admin_note"`
}
