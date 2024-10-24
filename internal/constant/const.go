package constant

type contextKey uint

const (
	AccessToken = "Access-Token"
	RefreshToken = "Refresh-Token"

	SuperUserRole string = "superuser"
	AdminRole string = "admin"
	CaptainRole string = "captain"

    UserIDContextKey contextKey = iota
	UserEmailContextKey
	RoleNameContextKey
	TeamIDContextKey

	JwtClaimsAttrUserID = "id"
	JwtClaimsAttrUserEmail = "email"
	JwtClaimsAttrRoleName = "role_name"
	JwtClaimsAttrTeamID = "team_id"
	JwtClaimsAttrTokenExpire = "token_exp"
)
