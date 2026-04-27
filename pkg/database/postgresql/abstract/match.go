package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type IMatchR interface {
	GetPastMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.MatchV2, error)
	GetMatchListByGameID(ctx context.Context, logger zerolog.Logger, gameID int64) ([]entities.MatchV2, error)
	GetMatchListByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Match, error)
	GetMatchListByTournamentId(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.Match, error)
	FetchMatches(logger zerolog.Logger, ctx context.Context, gameID int64) ([]entities.ShortMatch, error)
	GetMatchById(logger zerolog.Logger, ctx context.Context, matchId int64) (*entities.Match, error)
}

type IMatchRW interface {
	CreateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) (id int64, err error)
	CreateGameMatch(logger zerolog.Logger, ctx context.Context, match entities.GamesMatch, gameId int64, tx tx.ITx) (id int64, err error)
	UpdateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) error
	DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)
	DeleteOldGameMatches(ctx context.Context, logger zerolog.Logger, gameId int64, newMatchesIds []int, tx tx.ITx) error
	UpdateOldMatch(ctx context.Context, logger zerolog.Logger, gameId int64, match *entities.NewMatch, tx tx.ITx) error
	CreateNewMatch(ctx context.Context, logger zerolog.Logger, gameId int64, match *entities.NewMatch, tx tx.ITx) error
	DeleteTournamentMatches(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	RewriteMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entities.Match, ratings map[int64]entities.Rating) error
	RewriteTournamentMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entities.Match, ratings map[int64]entities.TournamentRating, tx tx.ITx) error
}
