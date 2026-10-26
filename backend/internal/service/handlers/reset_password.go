package handlers

type PaswordResetRequest struct {
	Code string `json:"login"`
}

type PasswordResetResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func ResetPasswordHandler() {

}
