package checker

type JsonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
