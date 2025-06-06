package team

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"math"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"strconv"
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

func (s *Service) GetTeam(ctx context.Context, teamID int64) (entity.GetTeamResponseV2, error) {
	logger := s.logger.With().Str("service", "GetTeam").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	var (
		teamResponse    entity.GetTeamResponseV2
		team            entity.TeamV2
		teamLeagues     []entity.LeagueShort
		captain         entity.User
		teamLeagueStats []entity.TeamLeagueStat
	)

	g, ctx := errgroup.WithContext(timeout)

	g.Go(func() error {
		var err error
		team, err = s.rdbOperations.GetTeamById(logger, timeout, teamID)
		return err
	})

	g.Go(func() error {
		var err error
		teamLeagues, err = s.rdbOperations.GetLeagueListByTeamId(logger, timeout, teamID)
		return err
	})

	g.Go(func() error {
		var err error
		captain, err = s.rdbOperations.GetCaptainByTeamId(logger, timeout, teamID)
		return err
	})

	if err := g.Wait(); err != nil {
		return entity.GetTeamResponseV2{}, err
	}

	fullPlayers, err := s.getPlayersAddStat(timeout, team.PlayersIds)
	if err != nil {
		return entity.GetTeamResponseV2{}, err
	}

	for _, l := range teamLeagues {
		teamLeagueStat := entity.TeamLeagueStat{}

		games, err := s.rdbOperations.GetTeamGamesInLeague(logger, timeout, teamID, l.ID)
		if err != nil {
			return entity.GetTeamResponseV2{}, err
		}

		var scoredGoals int = 0
		var concededGoals int = 0
		var points int = 0

		for _, game := range games {

			// если есть техническое поражение у игры,то матчи не смотрим,т.к. такой матч не создаётся
			if game.TechLooseTeamId != nil {
				if *game.TechLooseTeamId != int(teamID) {
					scoredGoals += constant.TechWinGoals
					concededGoals += constant.TechLooseGoals
					points += constant.WinPoints
					continue
				} else if *game.TechLooseTeamId == int(teamID) {
					concededGoals += constant.TechWinGoals
					scoredGoals += constant.TechWinGoals
					continue
				}
			}

			matches, err := s.rdbOperations.GetMatchListByGameID(timeout, logger, game.Id)
			if err != nil {
				return entity.GetTeamResponseV2{}, err
			}

			for _, m := range matches {
				if game.IsHomeGame {
					scoredGoals += m.ScoreTeam1
					concededGoals += m.ScoreTeam2
					if m.ScoreTeam1 > m.ScoreTeam2 {
						points += constant.WinPoints
					}
				} else {
					scoredGoals += m.ScoreTeam2
					concededGoals += m.ScoreTeam1
					if m.ScoreTeam2 > m.ScoreTeam1 {
						points += constant.WinPoints
					}
				}
			}
		}

		teamLeagueStat.League = l
		teamLeagueStat.Points = points
		teamLeagueStat.ScoreDifference = scoredGoals - concededGoals
		teamLeagueStat.GamesCount = len(games)
		teamLeagueStat.BestPlayer = findBestPlayer(fullPlayers, int(l.ID))

		teamLeagueStats = append(teamLeagueStats, teamLeagueStat)
	}

	teamResponse.ID = team.Id
	teamResponse.Name = team.Name
	teamResponse.ShortName = team.ShortName
	teamResponse.Avatar = team.Avatar
	teamResponse.CityId = team.CityId
	teamResponse.Captain = captain
	teamResponse.LeaguesStats = teamLeagueStats
	teamResponse.Players = fullPlayers

	return teamResponse, nil
}

func (s *Service) getPlayersAddStat(ctx context.Context, playersIds []int64) ([]entity.FullPlayer, error) {
	logger := s.logger.With().Str("service", "getPlayersAddStat").Logger()

	var fullPlayers []entity.FullPlayer

	for i := range playersIds {

		var fp entity.FullPlayer

		fullPlayer, err := s.playerSrv.Get(ctx, int(playersIds[i]))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get player.Get in team.GetTeam")
			return nil, err
		}

		fp.ID = fullPlayer.ID
		fp.Name = fullPlayer.Name
		fp.SecondName = fullPlayer.SecondName
		fp.LastName = fullPlayer.LastName
		fp.Avatar = fullPlayer.Avatar
		fp.ActivePlayer = fullPlayer.ActivePlayer
		fp.Deleted = fullPlayer.Deleted
		fp.CityID = fullPlayer.CityID
		fp.CityName = fullPlayer.CityName

		if fullPlayer.Leagues != nil {

			fp.Leagues = make([]entity.LeagueItem, len(fullPlayer.Leagues))

			for j, l := range fullPlayer.Leagues {
				fp.Leagues[j].ID = l.ID
				fp.Leagues[j].Name = l.Name
				fp.Leagues[j].Rating = l.Rating
				fp.Leagues[j].MatchesPlayed = l.MatchesPlayed
				fp.Leagues[j].GoalsScoredNumber = l.GoalsScoredNumber
				fp.Leagues[j].GoalsConcededNumber = l.GoalsConcededNumber
				fp.Leagues[j].GamesPlayedNumber = l.GamesPlayedNumber
				fp.Leagues[j].PercentageOfParticipation = l.PercentageOfParticipation

				if l.Teams != nil {

					fp.Leagues[j].Teams = make([]entity.TeamItem, len(l.Teams))

					for k, t := range l.Teams {
						fp.Leagues[j].Teams[k].ID = t.ID
						fp.Leagues[j].Teams[k].Name = t.Name
						fp.Leagues[j].Teams[k].ShortName = t.ShortName
						fp.Leagues[j].Teams[k].Avatar = t.Avatar
						fp.Leagues[j].Teams[k].CityID = t.CityID
						fp.Leagues[j].Teams[k].Leagues = t.Leagues
					}
				}
			}
		}

		fullPlayers = append(fullPlayers, fp)
	}

	return fullPlayers, nil
}

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

				extraPoints, err := s.rdbOperations.GetTeamExtraPointsCount(logger, ctx, team.ID, league.ID)
				if err != nil {
					logger.Error().Err(err).Msg("database error")
					return entity.GetTeamVsTeamTableResponse{Message: "database error"}, err
				}
				bodyItem.Score += extraPoints // Один раз за лигу считаем дополнительные очки команды

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

func findBestPlayer(players []entity.FullPlayer, leagueId int) entity.FullPlayer {
	var bestPlayer = entity.FullPlayer{}

	var maxRating = math.MinInt

	for _, player := range players {
		for _, lp := range player.Leagues {
			if lp.ID == leagueId && lp.Rating > maxRating {
				maxRating = lp.Rating
				bestPlayer = player
				break
			}
		}
	}

	return bestPlayer
}
