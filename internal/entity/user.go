package entity

type User struct {
	ID       int64  `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	Password []byte `json:"password" db:"password"`
	Role     *Role
	Team     *Team
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleName string `json:"roleName"`
	TeamID   string `json:"teamId,omitempty"`
}

type CreateUserResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct{
	Message string 	`json:"message"`
	ID 		int64	`json:"id"`
	Token 	string	`json:"token"`
	Role 	string	`json:"role"`
	TeamID	int64	`json:"teamId"`
}

type ChangePasswordRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
}

type CheckAuthRequest struct {
	UserID int64
	Token  string
}

type CheckAuthResponse struct {
	IsAuthenticated bool 	`json:"isAuthenticated"`
	Role 			string 	`json:"role"`
	TeamID 			int64 	`json:"teamId"`
}
