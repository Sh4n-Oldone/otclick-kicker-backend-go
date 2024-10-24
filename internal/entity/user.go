package entity

type User struct{
	ID int64 `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
	Password []byte `json:"password" db:"password"`
	Role *Role
	Team *Team
}

type CreateUserRequest struct{
	Email string	`json:"email"`
	Password string `json:"password"`
	RoleName string `json:"role_name"`
	TeamID string `json:"team_id,omitempty"`
}

type CreateUserResponse struct{
	Token string `json:"token"`
	Message string `json:"message"`
}

type LoginUserRequest struct{
	Email string	`json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct{
	Email string	`json:"email"`
	Password string `json:"password"`
	NewPassword string `json:"new_password"`
}
