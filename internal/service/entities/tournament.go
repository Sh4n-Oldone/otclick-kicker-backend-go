package entities

import "time"

type GetTournamentTypeListRequest struct {
	WithDeleted bool `json:"withDeleted"`
}

type CreateTournamentRequest struct {
	CityIDParam      string
	CityID           int64
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
	ID int64 `validate:"required,gt=0"`
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

type TournamentStage struct {
	ID           int64 `json:"id"`
	TournamentID int64 `json:"tournamentId"`
	IsFinished   bool  `json:"isFinished"`
}

type TournamentType struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type TournamentRule struct {
	Regular *Regular `json:"regular,omitempty"`
	PlayOff *PlayOff `json:"playOff,omitempty"`
}

type Regular struct {
	BestOf int64 `json:"bestOf" validate:"required,gt=0"`
}

type PlayOff struct {
	BestOf int64   `json:"bestOf" validate:"required,gt=0"`
	Looser *Looser `json:"looser,omitempty"`
}

type Looser struct {
	BestOf int64 `json:"bestOf" validate:"required,gt=0"`
}

type FinishStageRequest struct {
	ID       int64 `validate:"required,gt=0"`
	Finisher User
}

type NullableStage struct {
	ID           int64
	TournamentID *int64
	IsFinished   *bool
}

type TournamentStageItem struct {
	Stage TournamentStage  `json:"stage"`
	Games []TournamentGame `json:"games"`
}
