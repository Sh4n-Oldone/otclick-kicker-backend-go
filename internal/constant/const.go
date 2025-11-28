package constant

type contextKey uint

const (
	AccessToken = "Access-Token"

	SuperUserRole    string = "superuser"
	AdminRole        string = "admin"
	CaptainRole      string = "captain"
	TournamentMaster string = "tournament-master"
	CaptainRoleId    int    = 3

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

	RegularTournamentTypeID                  int64 = 1
	PlayoffTournamentTypeID                  int64 = 2
	RegularPlayoffTournamentTypeID           int64 = 3
	RegularPlayoffWithLooserTournamentTypeID int64 = 4
	RegularOneVsOneTournamentTypeID          int64 = 5

	SeparatorStageNumber string = "/"

	FirstStage = 1
)
