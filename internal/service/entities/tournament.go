package entities

import "time"

type GetTournamentTypeListRequest struct {
	WithDeleted bool `json:"withDeleted"`
}

type CreateTournamentRequest struct {
	CityID           *int64         `json:"cityId" validate:"omitempty,gt=0"`
	TournamentTypeID int64          `json:"typeId" validate:"required,gt=0"`
	SeasonID         int64          `json:"seasonId" validate:"required,gt=0"`
	Name             string         `json:"name" validate:"required,min=1,max=128"`
	TeamsIDs         []int64        `json:"teamsIds" validate:"omitempty,min=2"`
	PlayersIDs       []int64        `json:"playersIds" validate:"omitempty,min=2"`
	Rules            TournamentRule `json:"rules" validate:"required"`
	Creator          User
}

type UpdateTournamentRequest struct {
	ID               int64           `json:"id" validate:"required,gt=0"`
	CityID           *int64          `json:"cityId" validate:"omitempty,gt=0"`
	SeasonID         *int64          `json:"seasonId" validate:"omitempty,gt=0"`
	Name             *string         `json:"name" validate:"omitempty,min=1,max=128"`
	TeamsIDs         []int64         `json:"teamsIds" validate:"omitempty,min=2"`
	PlayersIDs       []int64         `json:"playersIds" validate:"omitempty,min=2"`
	Rules            *TournamentRule `json:"rules"`
	TournamentTypeID int64
	Executor         User
}

type DeleteTournamentRequest struct {
	ID       int64 `validate:"required,gt=0"`
	Executor User
}

type Tournament struct {
	ID       int64             `json:"id"`
	TypeID   int64             `json:"typeId"`
	Name     string            `json:"name"`
	Rules    TournamentRule    `json:"rules"`
	CityID   int64             `json:"cityId"`
	SeasonID int64             `json:"seasonId"`
	Stages   []TournamentStage `json:"stages"`
	TeamIDs  []int64           `json:"teamIds"`
}

type TournamentShort struct {
	ID       int64          `json:"id"`
	TypeID   int64          `json:"typeId"`
	Name     string         `json:"name"`
	Rules    TournamentRule `json:"rules"`
	CityID   int64          `json:"cityId"`
	SeasonID int64          `json:"seasonId"`
}

type TournamentStage struct {
	ID           int64  `json:"id"`
	TournamentID int64  `json:"tournamentId"`
	IsFinished   bool   `json:"isFinished"`
	Number       string `json:"number"`
}

type TournamentType struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type TournamentRule struct {
	Regular        *Regular        `json:"regular,omitempty"`
	PlayOff        *PlayOff        `json:"playOff,omitempty"`
	RegularPlayoff *RegularPlayoff `json:"regularPlayoff,omitempty"`
}

type Regular struct {
	BestOf int64 `json:"bestOf" validate:"required,gt=0"`
}

type PlayOff struct {
	Stages map[int64]BestOf `json:"stages" validate:"required=true,min=1"`
}

type RegularPlayoff struct {
	Regular                  Regular `json:"regular" validate:"required=true"`
	PlayOff                  PlayOff `json:"playOff" validate:"required=true"`
	PlayoffTeamsCountOnStart int32   `json:"playoffTeamsCountOnStart" validate:"required=true,gt=0"`
}
type BestOf struct {
	Bo int64 `json:"bo" validate:"required,gt=0"`
}

type FinishStageRequest struct {
	ID       int64 `validate:"required,gt=0"`
	Finisher User
}

type NullableStage struct {
	ID           int64
	TournamentID *int64
	IsFinished   *bool
	Number       *string
}

type TournamentStageItem struct {
	Stage TournamentStage  `json:"stage"`
	Games []TournamentGame `json:"games"`
}

type GetTournamentListRequest struct {
	CityID           *int64 `validate:"omitempty,gt=0"`
	SeasonID         *int64 `validate:"omitempty,gt=0"`
	TournamentID     *int64 `validate:"omitempty,gt=0"`
	TournamentTypeID *int64 `validate:"omitempty,gt=0"`
}

type TournamentItem struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Rating int    `json:"rating"`

	// tournament stats
	MatchesPlayed             int     `json:"matchesPlayed"`
	GoalsScoredNumber         int     `json:"goalsScoredNumber"`
	GoalsConcededNumber       int     `json:"goalsConcededNumber"`
	GamesPlayedNumber         int     `json:"gamesPlayedNumber"`
	PercentageOfParticipation float32 `json:"percentageOfParticipation"`

	Teams []TeamItem `json:"teams,omitempty"`
}

type RecalcTournamentRequest struct {
	Id int64 `validate:"required,gt=0"`
}

type StartNextStageRequest struct {
	TournamentID int64   `validate:"required,gt=0"`
	TeamsIds     []int64 `json:"teamsIds" validate:"omitempty,min=2"`
	Executor     User
}

type StageStat struct {
	Stage     TournamentStage
	Number    int64
	GameStats []GameStat
	Winners   []int64
	Losers    []int64
	BestOf    int64
}
type TeamTournamentIdRequest struct {
	TeamId       int64 `json:"teamId"`
	TournamentId int64 `json:"tournamentId"`
}

type ExtraPointsTournament struct {
	Id           int64  `json:"id"`
	TeamId       int64  `json:"teamId"`
	TournamentId int64  `json:"tournamentId"`
	Reason       string `json:"reason"`
	Points       int64  `json:"points"`
}

type CreateExtraPointsTournamentRequest struct {
	TeamId       int64  `json:"teamId" validate:"required,gt=0"`
	TournamentId int64  `json:"tournamentId" validate:"required,gt=0"`
	Reason       string `json:"reason" validate:"required"`
	Points       int64  `json:"points"`
	Role         string
	UserId       int64
}

type UpdateExtraPointsTournamentRequest struct {
	Id           int64   `json:"id" validate:"required,gt=0"`
	TeamId       *int64  `json:"teamId" validate:"omitempty,gt=0"`
	TournamentId *int64  `json:"tournamentId" validate:"omitempty,gt=0"`
	Reason       *string `json:"reason"`
	Points       *int64  `json:"points"`
	Role         string
	UserId       int64
}

type DeleteExtraPointsRequest struct {
	Id     int64 `json:"id" validate:"required,gt=0"`
	UserId int64
	Role   string
}

type TournamentBrackets struct {
	TeamsOfWinnerBracket []*int64
	TeamsOfLoserBracket  []*int64
}

type EmptyRequest struct{}
