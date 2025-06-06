package constant

type contextKey uint

const (
	AccessToken = "Access-Token"

	SuperUserRole string = "superuser"
	AdminRole     string = "admin"
	CaptainRole   string = "captain"
	CaptainRoleId int    = 3

	UserIDContextKey contextKey = iota
	UserEmailContextKey
	RoleNameContextKey
	TeamIDContextKey

	JwtClaimsAttrUserID      = "id"
	JwtClaimsAttrUserEmail   = "email"
	JwtClaimsAttrRoleName    = "role_name"
	JwtClaimsAttrTeamID      = "team_id"
	JwtClaimsAttrTokenExpire = "token_exp"

	DefaultRating int = 1000

	WinPoints      int = 2
	TechWinGoals   int = 42
	TechLooseGoals int = 30
)
