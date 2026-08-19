package auth

type registerRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type registerResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type loginRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type loginResponse struct {
	Email *string `json:"email"`
}
