package team

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

// GetTeam {id}
// GetTeams
// GetTeamsByCity {city_id}
// GetTeamsByLeague {league_id}
// GetTeamVsTeamTable {city_id}

// Create
// Update
// Delete {id}

// AddPlayerIntoTeam
// RemovePlayerFromTeam
// //////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Service) GetTeam(ctx context.Context, teamID int64) (entity.GetTeamResponse, error) {
	logger := s.logger.With().Interface("service", "GetTeam").Logger()

	response, err := s.rdbOperations.GetTeam(logger, ctx, teamID)
	if err != nil {
		return response, err
	}

	if err := s.addPlayersProperties(ctx, response.Players); err != nil {
		logger.Error().Err(err).Msg("Failed to add player properties")
		return response, err
	}

	return response, nil
}

func (s *Service) addPlayersProperties(ctx context.Context, players []entity.PlayerGetTeam) error {
	logger := s.logger.With().Interface("service", "addPlayersProperties").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	for i := range players {
		p := &players[i] // Создаём указатель на текущий элемент массива

		pastMatches, err := s.rdbOperations.GetPastMatchesByPlayerID(logger, timeout, p.ID)
		if err != nil {
			return err
		}

		propertyCounting(p, pastMatches)
	}

	return nil
}

func propertyCounting(player *entity.PlayerGetTeam, pastMatches []entities.Match) {
	playersGames := make(map[int]struct{})

	var (
		goalsScoredNumber   int
		goalsConcededNumber int
	)

	for _, match := range pastMatches {
		playersGames[match.GameID] = struct{}{}

		if match.Player1Team1ID == player.ID || (match.Player2Team1ID != nil && *match.Player2Team1ID == player.ID) {
			goalsScoredNumber += match.ScoreTeam1
			goalsConcededNumber += match.ScoreTeam2
		}
		if match.Player1Team2ID == player.ID || (match.Player2Team2ID != nil && *match.Player2Team2ID == player.ID) {
			goalsScoredNumber += match.ScoreTeam2
			goalsConcededNumber += match.ScoreTeam1
		}
	}

	player.MatchesPlayed = len(pastMatches)
	player.GamesPlayedNumber = len(playersGames)
	player.GoalsScoredNumber = goalsScoredNumber
	player.GoalsConcededNumber = goalsConcededNumber
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Service) GetTeams(ctx context.Context, cityId int64, onlyFree bool) ([]entity.TeamShort, error) {
	logger := s.logger.With().Interface("service", "GetTeams").Logger()

	teams, err := s.rdbOperations.GetTeams(logger, ctx, cityId, onlyFree)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamsByCity(ctx context.Context, onlyFree bool, cityID int64) ([]entity.TeamShort, error) {
	logger := s.logger.With().Interface("service", "GetTeamsByCity").Logger()

	teams, err := s.rdbOperations.GetTeamsByCity(logger, ctx, onlyFree, cityID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamsByLeague(ctx context.Context, leagueID int64) ([]entity.TeamByLeague, error) {
	logger := s.logger.With().Interface("service", "GetTeamsByLeague").Logger()

	teams, err := s.rdbOperations.GetTeamsByLeague(logger, ctx, leagueID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

// ///////////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////////

func (s *Service) GetTeamVsTeamTable(ctx context.Context, cityID, seasonID int64) (entity.GetTeamVsTeamTableResponse, error) {
	logger := s.logger.With().Interface("service", "GetTeamVsTeamTable").Logger()
	leagues, err := s.rdbOperations.FetchLeagues(logger, ctx, cityID, seasonID)

	if err != nil {
		logger.Error().Err(err).Msg("error GetTeamVsTeamTable")
		return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
	}

	if len(leagues) == 0 {
		logger.Error().Err(err).Msg("No leagues found")
		return entity.GetTeamVsTeamTableResponse{Message: "No leagues found"}, error_templates.New("No leagues found", errors.New("No leagues found"), codes.NotFound, http.StatusNotFound)
	}

	var data []entity.Data
	//////////////////////////////////
	for _, league := range leagues { // Проходим по лигам нужного города
		var dataItem entity.Data
		dataItem.LeagueID = league.ID
		dataItem.LeagueName = league.Name

		dataItem.Table.Columns = append(dataItem.Table.Columns, entity.Column{Uid: "teamShortName", Name: "Команда"})

		teams, err := s.rdbOperations.FetchTeams(logger, ctx, league.ID)
		if err != nil {
			logger.Error().Err(err).Msg("error get teams")
			return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
		}
		if len(teams) == 0 {
			continue
		}
		noGames, err := s.rdbOperations.TeamsHaveNoGames(logger, ctx, teams, seasonID)
		if err != nil {
			logger.Error().Err(err).Msg("database error")
			return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
		}
		if noGames { // если нет игр в лиге
			for _, team := range teams { // голы заполняем нулями
				var bodyItem entity.Body

				bodyItem.Id = team.ID
				bodyItem.TeamShortName = team.ShortName
				bodyItem.Score = 0
				bodyItem.DifferenceInScore = 0
				bodyItem.GamesPlayed = 0
				bodyItem.GamesToPlay = int64((len(teams) - 1) * 2)
				for _, t := range teams { // и отображаем нулевой счет
					var cell entity.TableCell

					cell.Game1ID = 0
					cell.Game2ID = 0
					cell.Score1 = "0:0"
					cell.Score2 = "0:0"

					bodyItem.TableCell[t.ShortName] = cell
				}
				dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)
			}
			continue
		}
		//////////////////////////////////
		for _, team := range teams { // Проходим по командам текущей лиги
			var bodyItem entity.Body
			bodyItem.TableCell = make(map[string]entity.TableCell, 0)

			bodyItem.Id = team.ID
			bodyItem.TeamShortName = team.ShortName
			bodyItem.Score = 0
			bodyItem.DifferenceInScore = 0
			bodyItem.GamesPlayed = 0
			bodyItem.GamesToPlay = int64((len(teams) - 1) * 2)
			//////////////////////////////////
			for _, t := range teams { // Проходим по командам-соперникам
				var cell entity.TableCell

				cell.Game1ID = 0
				cell.Game2ID = 0
				cell.Score1 = "0:0"
				cell.Score2 = "0:0"

				if t.ID == team.ID {
					bodyItem.TableCell[t.ShortName] = cell
					continue //команда сама с собой не играет
				}
				gamesHome, err := s.rdbOperations.FetchPastGames(logger, ctx, team.ID, t.ID, cityID, league.ID)
				if err != nil {
					logger.Error().Err(err).Msg("database error")
					return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
				}
				gamesOut, err := s.rdbOperations.FetchPastGames(logger, ctx, t.ID, team.ID, cityID, league.ID)
				if err != nil {
					logger.Error().Err(err).Msg("database error")
					return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
				}

				if len(gamesHome) == 0 && len(gamesOut) == 0 { //если нет ни домашних ни выездных игр
					bodyItem.TableCell[t.ShortName] = cell
					continue
				}
				if len(gamesHome) != 0 { // если есть дом
					cell.Game1ID = int64(gamesHome[0].ID)
				}
				if len(gamesOut) != 0 { // если есть выезд
					cell.Game2ID = int64(gamesOut[0].ID)
				}

				var match1Team1Score int64 = 0
				var match1Team2Score int64 = 0

				var match2Team1Score int64 = 0
				var match2Team2Score int64 = 0

				if len(gamesHome) != 0 {
					cell.Game1ID = int64(gamesHome[0].ID)
					if gamesHome[0].TechLooseTeamID != nil && *gamesHome[0].TechLooseTeamID == team.ID { // Если команда с тех.проигрышем(team.ID) то "30:42" (эта команда проиграла)
						cell.Score1 = "30:42"
						bodyItem.DifferenceInScore -= 12 // Разницу учитываем
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else if gamesHome[0].TechLooseTeamID != nil && *gamesHome[0].TechLooseTeamID == t.ID { // Если противника команда с тех.проигрышем(t.ID) то "42:30" (противник проиграл)
						cell.Score1 = "42:30"
						bodyItem.DifferenceInScore += 12 // Разницу учитываем
						bodyItem.Score += 2              // Добавим себе 2 очка(в целом по игре), если команда противника с тех.проигрышем
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else {
						gamesHomeMatches, err := s.rdbOperations.FetchMatches(logger, ctx, cell.Game1ID)
						if err != nil {
							logger.Error().Err(err).Msg("database error")
							return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
						}
						//////////////////////////////////
						for _, match := range gamesHomeMatches { // Проходим по домашним матчам
							if match.Team1ID == int(team.ID) {
								match1Team1Score += int64(match.ScoreTeam1)
								match1Team2Score += int64(match.ScoreTeam2)
							} else {
								match1Team1Score += int64(match.ScoreTeam2)
								match1Team2Score += int64(match.ScoreTeam1)
							}
						}
						cell.Score1 = strconv.FormatInt(match1Team1Score, 10) + ":" + strconv.FormatInt(match1Team2Score, 10)

						if len(gamesHomeMatches) != 0 {
							bodyItem.GamesPlayed += 1
							bodyItem.GamesToPlay -= 1
						}
					}
				}

				if len(gamesOut) != 0 {
					cell.Game2ID = int64(gamesOut[0].ID)
					if gamesOut[0].TechLooseTeamID != nil && *gamesOut[0].TechLooseTeamID == team.ID { // Если команда с тех.проигрышем(team.ID) то "30:42" (эта команда проиграла)
						cell.Score2 = "30:42"
						bodyItem.DifferenceInScore -= 12 // Разницу учитываем
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else if gamesOut[0].TechLooseTeamID != nil && *gamesOut[0].TechLooseTeamID == t.ID { // Если противника команда с тех.проигрышем(t.ID) то "42:30" (противник проиграл)
						cell.Score2 = "42:30"
						bodyItem.DifferenceInScore += 12 // Разницу учитываем
						bodyItem.Score += 2              // Добавим себе 2 очка(в целом по игре), если команда противника с тех.проигрышем
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else {
						gamesOutMatches, err := s.rdbOperations.FetchMatches(logger, ctx, cell.Game2ID)
						if err != nil {
							logger.Error().Err(err).Msg("database error")
							return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
						}
						//////////////////////////////////
						for _, match := range gamesOutMatches { // Проходим по выездным матчам
							if match.Team2ID == int(t.ID) {
								match2Team1Score += int64(match.ScoreTeam1)
								match2Team2Score += int64(match.ScoreTeam2)
							} else {
								match2Team1Score += int64(match.ScoreTeam2)
								match2Team2Score += int64(match.ScoreTeam1)
							}
						}
						cell.Score2 = strconv.FormatInt(match2Team1Score, 10) + ":" + strconv.FormatInt(match2Team2Score, 10)

						if len(gamesOutMatches) != 0 {
							bodyItem.GamesPlayed += 1
							bodyItem.GamesToPlay -= 1
						}
					}
				}

				bodyItem.Score = resumScore(bodyItem.Score, match1Team1Score, match1Team2Score)
				bodyItem.Score = resumScore(bodyItem.Score, match2Team1Score, match2Team2Score)
				bodyItem.DifferenceInScore += (match1Team1Score - match1Team2Score + match2Team1Score - match2Team2Score)
				bodyItem.TableCell[t.ShortName] = cell // Выставили ячейку со счетом
			}
			dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)

		}
		dataItem.Table.Columns = append(dataItem.Table.Columns, entity.Column{Uid: "score", Name: "Очки"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entity.Column{Uid: "differenceInScore", Name: "+/-"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entity.Column{Uid: "gamesPlayed", Name: "Игры"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entity.Column{Uid: "gamesToPlay", Name: "Осталось"})

		data = append(data, dataItem)

	}

	var response entity.GetTeamVsTeamTableResponse
	response.Data = data
	response.Message = "OK"
	return response, nil
}

func resumScore(score, team1Score, team2Score int64) int64 {
	if team1Score < 1 && team2Score < 1 {
		return score
	}

	if team1Score == team2Score {
		return score // both team got +2 instead?
	}

	if team1Score > team2Score {
		return score + 2
	}

	return score // -2 ?
}

// ///////////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////////
func (s *Service) Create(ctx context.Context, team entity.CreateTeamRequest) (int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateTeam(logger, ctx, team)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, UpdateTeamRequest entity.UpdateTeamRequest) (bool, error) {
	logger := s.logger.With().Interface("service", "Update").Logger()

	res, err := s.rwdbOperations.UpdateTeam(logger, ctx, UpdateTeamRequest)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (bool, error) {
	logger := s.logger.With().Interface("service", "Delete").Logger()

	res, err := s.rwdbOperations.DeleteTeam(logger, ctx, id)
	if err != nil {
		return res, err
	}

	return res, nil
}

// ///////////////////////////////////////////////////////////////////////////////////
func (s *Service) AddPlayerIntoTeam(ctx context.Context, playerID, teamID int64) (bool, error) {
	logger := s.logger.With().Interface("service", "AddPlayerIntoTeam").Logger()

	res, err := s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, playerID, teamID)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) RemovePlayerFromTeam(ctx context.Context, playerID, teamID int64) (bool, error) {
	logger := s.logger.With().Interface("service", "RemovePlayerFromTeam").Logger()

	res, err := s.rwdbOperations.RemovePlayerFromTeam(logger, ctx, playerID, teamID)
	if err != nil {
		return res, err
	}

	return res, nil
}
