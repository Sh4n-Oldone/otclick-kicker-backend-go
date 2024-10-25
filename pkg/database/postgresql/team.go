package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"

	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
)

/* func (db *RDBOperation) GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64) (entity.GetTeamResponse, error) {
		rows, err := db.db.Query(ctx, "queryGetTeamPlayers", teamID)
	   	if err != nil {
	   		logger.Error().Err(err).Msg("failed to GetTeam")
	   		return entity.GetTeamResponse{}, DecodeDatabaseError(err)
	   	}

	response := entity.GetTeamResponse{}
	return response, nil
} */

// func (db *RWDBOperation) GetTeams(logger zerolog.Logger, ctx context.Context, team entity.Team) (int64, error) {
// 	var id int64

// 	err := db.db.QueryRow(ctx, queryGetTeams,
// 		team.ID,
// 	).Scan(&id)
// 	if err != nil {
// 		logger.Error().Err(err).Msg("failed to GetTeams")
// 		return 0, DecodeDatabaseError(err)
// 	}

// 	return id, nil
// }

// func (db *RWDBOperation) GetTeamsByCity(logger zerolog.Logger, ctx context.Context, team entity.Team) (int64, error) {
// 	var id int64

// 	err := db.db.QueryRow(ctx, queryGetTeamsByCity,
// 		team.ID,
// 	).Scan(&id)
// 	if err != nil {
// 		logger.Error().Err(err).Msg("failed to GetTeamsByCity")
// 		return 0, DecodeDatabaseError(err)
// 	}

// 	return id, nil
// }

// func (db *RWDBOperation) GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, team entity.Team) (int64, error) {
// 	var id int64

// 	err := db.db.QueryRow(ctx, queryGetTeamsByLeague,
// 		team.ID,
// 	).Scan(&id)
// 	if err != nil {
// 		logger.Error().Err(err).Msg("failed to GetTeamsByLeague")
// 		return 0, DecodeDatabaseError(err)
// 	}

// 	return id, nil
// }

// func (db *RWDBOperation) GetTeamVsTeamTable(logger zerolog.Logger, ctx context.Context, team entity.Team) (int64, error) {
// 	var id int64

// 	err := db.db.QueryRow(ctx, queryGetTeamVsTeamTable,
// 		team.ID,
// 	).Scan(&id)
// 	if err != nil {
// 		logger.Error().Err(err).Msg("failed to GetTeamVsTeamTable")
// 		return 0, DecodeDatabaseError(err)
// 	}

// 	return id, nil
// }

// //////////////////////////////////////////////////////////////////////////////////
func (db *RWDBOperation) CreateTeam(logger zerolog.Logger, ctx context.Context, team entity.CreateTeamRequest) (int64, error) {
	var id int64
	const queryCreateTeam = `INSERT INTO teams (name, short_name, avatar, city_id, league_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := db.db.QueryRow(ctx, queryCreateTeam,
		team.Name,
		team.ShortName,
		team.Avatar,
		team.CityId,
		team.LeagueID,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Team record")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdateTeam(logger zerolog.Logger, ctx context.Context, team entity.UpdateTeamRequest) (bool, error) {
	var fields []string
	var values []interface{}
	index := 1

	// Проверяем и добавляем поля для обновления
	if team.Name != nil {
		fields = append(fields, fmt.Sprintf("name = $%d", index))
		values = append(values, *team.Name)
		index++
	}
	if team.ShortName != nil {
		fields = append(fields, fmt.Sprintf("short_name = $%d", index))
		values = append(values, *team.ShortName)
		index++
	}
	if team.Avatar != nil {
		fields = append(fields, fmt.Sprintf("avatar = $%d", index))
		values = append(values, *team.Avatar)
		index++
	}
	if team.CityId != nil {
		fields = append(fields, fmt.Sprintf("city_id = $%d", index))
		values = append(values, *team.CityId)
		index++
	}
	if team.LeagueID != nil {
		fields = append(fields, fmt.Sprintf("league_id = $%d", index))
		values = append(values, *team.LeagueID)
		index++
	}

	// Если нет полей для обновления
	if len(fields) == 0 {
		logger.Warn().Msg("No fields to update")
		return false, nil
	}

	// Добавляем ID команды как последний параметр
	fieldsQuery := strings.Join(fields, ", ")
	queryUpdateTeam := fmt.Sprintf("UPDATE teams SET %s WHERE id = $%d", fieldsQuery, index)
	values = append(values, team.ID)

	// Выполняем запрос
	result, err := db.db.Exec(ctx, queryUpdateTeam, values...)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update Team record")
		return false, DecodeDatabaseError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Error().Err(err).Msg("failed to get affected rows")
		return false, stderr.New("Failed to Update Team, it does not exist")
	}

	return true, nil
}

func (db *RWDBOperation) DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64) (bool, error) {
	const queryDeleteTeam = `DELETE FROM teams WHERE id = $1`
	result, err := db.db.Exec(ctx, queryDeleteTeam, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete Team record")
		return false, DecodeDatabaseError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Error().Err(err).Msg("failed to get affected rows")
		return false, stderr.New("Failed to Delete Team, it does not exist")
	}

	return true, nil
}

// ////////////////////////////////////////////////////////////////////////////////
func (db *RWDBOperation) AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error) {
	var exists int
	const query1 = "SELECT 1 FROM players_teams_links WHERE player_id = $1 AND team_id = $2"
	err := db.db.QueryRow(ctx, query1, playerID, teamID).Scan(&exists)
	if err == nil {
		logger.Error().Err(err).Msg("Player already in that team!")
		return false, stderr.New("Player already in that team!")
	}
	if err != pgx.ErrNoRows {
		logger.Error().Err(err).Msg("Team or player not found")
		return false, DecodeDatabaseError(err)
	}

	const query2 = "INSERT INTO players_teams_links(player_id, team_id) VALUES($1, $2)"
	tag, err := db.db.Exec(ctx, query2, playerID, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to AddPlayerIntoTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		logger.Error().Msg("No rows affected, failed to insert player into team")
		return false, stderr.New("Failed to add player into team")
	}

	return true, nil
}

func (db *RWDBOperation) RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error) {
	var exists int
	const query1 = "SELECT 1 FROM players_teams_links WHERE player_id = $1 AND team_id = $2"
	err := db.db.QueryRow(ctx, query1, playerID, teamID).Scan(&exists)
	if err == pgx.ErrNoRows {
		logger.Error().Msg("Player not in that team!")
		return false, stderr.New("Player not in that team!")
	}
	if err != nil {
		logger.Error().Err(err).Msg("Team or player not found")
		return false, DecodeDatabaseError(err)
	}

	const query2 = "DELETE FROM players_teams_links WHERE player_id = $1 AND team_id = $2"
	tag, err := db.db.Exec(ctx, query2, playerID, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to RemovePlayerFromTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		logger.Error().Msg("No rows affected, failed to remove player from team")
		return false, stderr.New("failed to remove player from team")
	}

	return true, nil
}
