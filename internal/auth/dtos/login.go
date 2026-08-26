package dtos

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginOrgRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type CreateUserRequest struct {
	Fullname string `json:"fullname"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}
type CreateOrganizationRequest struct {
	// required
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Email    string `json:"email"     validate:"required,email"`
	// optionals
	Description string `json:"description"     validate:"omitempty,min=5,max=100"`
	Website     string `json:"website"     validate:"omitempty,min=5,max=100"`
	Logo        string `json:"logo"     validate:"omitempty,min=5,max=100"`
}
