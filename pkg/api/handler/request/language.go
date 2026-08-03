package request

// LanguageRequest is the admin create/update payload for a selectable language.
// Pointers on IsActive/SortOrder let update distinguish "omitted" from zero.
type LanguageRequest struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	SortOrder  *int   `json:"sort_order"`
	IsActive   *bool  `json:"is_active"`
}
