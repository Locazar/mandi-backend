package response

// LanguageItem is one selectable language as the client expects it.
type LanguageItem struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
}
