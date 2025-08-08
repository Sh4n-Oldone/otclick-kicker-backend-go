package entities

type User struct {
	ID       int64  `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	Password []byte `json:"password,omitempty" db:"password"`
	Role     *Role  `json:"role,omitempty"`
	Team     *Team  `json:"team,omitempty"`
}

type TournamentMaster struct {
	User User `json:"user"`
	City City `json:"city"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleName string `json:"roleName"`
	TeamID   int64  `json:"teamId,omitempty"`
}

type CreateTournamentMasterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6,lte=64"`
	RoleName string `json:"roleName" validate:"required,eq=tournament-master"`
	CityID   int64  `json:"cityId" validate:"required,gt=0"`
}

type CreateUserResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type UpdateTournamentMasterRequest struct {
	UserID int64  `json:"userId" validate:"required,gt=0"`
	CityID *int64 `json:"cityId" validate:"omitempty,gt=0"`
}

type GetTournamentMastersResponse struct {
	Masters []TournamentMaster `json:"masters"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
	Token   string `json:"token"`
	Role    string `json:"role"`
	TeamID  int64  `json:"teamId"`
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
	IsAuthenticated bool   `json:"isAuthenticated"`
	Role            string `json:"role"`
	TeamID          int64  `json:"teamId"`
}
