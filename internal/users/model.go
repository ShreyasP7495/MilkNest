package users

type UpdateProfileRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=1,max=120"`
	Email string `json:"email" validate:"omitempty,email"`
}
