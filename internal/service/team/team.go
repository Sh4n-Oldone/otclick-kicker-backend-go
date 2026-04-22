package team

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/convert"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers/pointer"
)

func (s *Service) GetTeam(ctx context.Context, teamID int64) (entities.GetTeamResponseV2, error) {
	logger := s.logger.With().Str("service", "GetTeam").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	var (
		teamResponse    entities.GetTeamResponseV2
		team            entities.TeamV2
		teamLeagues     []entities.LeagueShort
		captain         entities.User
		teamLeagueStats []entities.TeamLeagueStat
	)

	g, ctx := errgroup.WithContext(timeout)

	g.Go(func() error {
		var err error
		team, err = s.rdbOperations.GetTeamById(logger, timeout, teamID, nil)
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
		return entities.GetTeamResponseV2{}, err
	}

	fullPlayers, err := s.getPlayersAddStat(timeout, team.PlayersIds)
	if err != nil {
		return entities.GetTeamResponseV2{}, err
	}

	for _, l := range teamLeagues {
		teamLeagueStat := entities.TeamLeagueStat{}

		tiebreak := false // tiebreak игры не учитываем
		games, err := s.rdbOperations.GetTeamGamesInLeague(logger, timeout, teamID, l.ID, &tiebreak)
		if err != nil {
			return entities.GetTeamResponseV2{}, err
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
					scoredGoals += constant.TechLooseGoals
					concededGoals += constant.TechWinGoals
					continue
				}
			}

			matches, err := s.rdbOperations.GetMatchListByGameID(timeout, logger, int64(game.Id))
			if err != nil {
				return entities.GetTeamResponseV2{}, err
			}

			for _, m := range matches {
				// if game.IsHomeGame {
				if m.Team1ID == int(teamID) {
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

func (s *Service) getPlayersAddStat(ctx context.Context, playersIds []int64) ([]entities.FullPlayer, error) {
	logger := s.logger.With().Str("service", "getPlayersAddStat").Logger()

	var fullPlayers []entities.FullPlayer

	for i := range playersIds {

		var fp entities.FullPlayer

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

			fp.Leagues = make([]entities.LeagueItem, len(fullPlayer.Leagues))

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

					fp.Leagues[j].Teams = make([]entities.TeamItem, len(l.Teams))

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

func (s *Service) GetTeams(ctx context.Context, cityId int64, onlyFree bool) ([]entities.TeamShort, error) {
	logger := s.logger.With().Interface("service", "GetTeams").Logger()

	teams, err := s.rdbOperations.GetTeams(logger, ctx, cityId, onlyFree)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamsByCity(ctx context.Context, onlyFree bool, cityID int64) ([]entities.TeamShort, error) {
	logger := s.logger.With().Interface("service", "GetTeamsByCity").Logger()

	teams, err := s.rdbOperations.GetTeamsByCity(logger, ctx, onlyFree, cityID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamsByLeague(ctx context.Context, leagueID int64) ([]entities.TeamByLeague, error) {
	logger := s.logger.With().Interface("service", "GetTeamsByLeague").Logger()

	teams, err := s.rdbOperations.GetTeamsByLeague(logger, ctx, leagueID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamVsTeamTable(ctx context.Context, cityID, seasonID int64) (entities.GetTeamVsTeamTableResponse, error) {
	logger := s.logger.With().Interface("service", "GetTeamVsTeamTable").Logger()
	leagues, err := s.rdbOperations.FetchLeagues(logger, ctx, cityID, seasonID)

	if err != nil {
		logger.Error().Err(err).Msg("error GetTeamVsTeamTable")
		return entities.GetTeamVsTeamTableResponse{}, err
	}

	if len(leagues) == 0 {
		logger.Error().Err(err).Msg("No leagues found")
		return entities.GetTeamVsTeamTableResponse{}, error_templates.New("No leagues found", errors.New("No leagues found"), codes.NotFound, http.StatusNotFound)
	}

	var data []entities.Data
	//////////////////////////////////
	for _, league := range leagues { // Проходим по лигам нужного города
		var dataItem entities.Data
		dataItem.LeagueID = league.ID
		dataItem.LeagueName = league.Name

		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "teamShortName", Name: "Команда"})

		teams, err := s.rdbOperations.FetchTeams(logger, ctx, league.ID)
		if err != nil {
			logger.Error().Err(err).Msg("error get teams")
			return entities.GetTeamVsTeamTableResponse{}, err
		}
		if len(teams) == 0 {
			continue
		}
		noGames, err := s.rdbOperations.TeamsHaveNoGames(logger, ctx, teams, seasonID)
		if err != nil {
			logger.Error().Err(err).Msg("database error")
			return entities.GetTeamVsTeamTableResponse{}, err
		}
		if noGames { // если нет игр в лиге
			for _, team := range teams { // голы заполняем нулями
				var bodyItem entities.Body

				bodyItem.Id = team.ID
				bodyItem.TeamShortName = team.ShortName
				bodyItem.Score = 0
				bodyItem.DifferenceInScore = 0
				bodyItem.GamesPlayed = 0
				bodyItem.GamesToPlay = int64((len(teams) - 1) * 2)
				for _, t := range teams { // и отображаем нулевой счет
					var cell entities.TableCell

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
			var bodyItem entities.Body
			bodyItem.TableCell = make(map[string]entities.TableCell, 0)

			bodyItem.Id = team.ID
			bodyItem.TeamShortName = team.ShortName
			bodyItem.Score = 0
			bodyItem.DifferenceInScore = 0
			bodyItem.GamesPlayed = 0
			bodyItem.GamesToPlay = int64((len(teams) - 1) * 2)
			//////////////////////////////////
			for _, t := range teams { // Проходим по командам-соперникам
				var cell entities.TableCell

				cell.Game1ID = 0
				cell.Game2ID = 0
				cell.Score1 = "0:0"
				cell.Score2 = "0:0"

				if t.ID == team.ID {
					bodyItem.TableCell[t.ShortName] = cell
					continue //команда сама с собой не играет
				}

				tiebreak := false // tiebreak игры не учитываем

				gamesHome, err := s.rdbOperations.FetchPastGames(logger, ctx, team.ID, t.ID, cityID, league.ID, &tiebreak)
				if err != nil {
					logger.Error().Err(err).Msg("database error")
					return entities.GetTeamVsTeamTableResponse{}, err
				}
				gamesOut, err := s.rdbOperations.FetchPastGames(logger, ctx, t.ID, team.ID, cityID, league.ID, &tiebreak)
				if err != nil {
					logger.Error().Err(err).Msg("database error")
					return entities.GetTeamVsTeamTableResponse{}, err
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
							return entities.GetTeamVsTeamTableResponse{}, err
						}
						//////////////////////////////////
						for _, match := range gamesHomeMatches { // Проходим по домашним матчам
							if match.Team1ID == team.ID {
								match1Team1Score += match.ScoreTeam1
								match1Team2Score += match.ScoreTeam2
							} else {
								match1Team1Score += match.ScoreTeam2
								match1Team2Score += match.ScoreTeam1
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
							return entities.GetTeamVsTeamTableResponse{}, err
						}
						//////////////////////////////////
						for _, match := range gamesOutMatches { // Проходим по выездным матчам
							if match.Team2ID == t.ID {
								match2Team1Score += match.ScoreTeam1
								match2Team2Score += match.ScoreTeam2
							} else {
								match2Team1Score += match.ScoreTeam2
								match2Team2Score += match.ScoreTeam1
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
					return entities.GetTeamVsTeamTableResponse{}, err
				}
				bodyItem.Score += extraPoints // Один раз за лигу считаем дополнительные очки команды

				bodyItem.DifferenceInScore += (match1Team1Score - match1Team2Score + match2Team1Score - match2Team2Score)
				bodyItem.TableCell[t.ShortName] = cell // Выставили ячейку со счетом
			} // команды соперников

			dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)

		} // команды текущей лиги
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "score", Name: "Очки"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "differenceInScore", Name: "+/-"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "gamesPlayed", Name: "Игры"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "gamesToPlay", Name: "Осталось"})

		// добавили игры tiebreak текущей лиги
		gamesTiebreak, err := s.rdbOperations.FetchPastGamesTiebreak(logger, ctx, league.ID)
		if err != nil {
			logger.Error().Err(err).Msg("database error")
			return entities.GetTeamVsTeamTableResponse{}, err
		}
		dataItem.GamesTiebreak = gamesTiebreak

		data = append(data, dataItem)

	} // лиги нужного города

	var response entities.GetTeamVsTeamTableResponse
	response.Data = data
	response.Message = "OK"
	return response, nil
}

func (s *Service) GetTournamentTeamVsTeamTable(ctx context.Context, cityID, seasonID int64) (entities.GetTournamentTeamVsTeamTableResponse, error) {
	logger := s.logger.With().Interface("service", "GetTournamentTeamVsTeamTable").Logger()
	typeOneVsOne := constant.RegularOneVsOneTournamentTypeID
	tournamentReq := entities.GetTournamentListRequest{CityID: &cityID, SeasonID: &seasonID, TournamentID: nil, TournamentTypeID: &typeOneVsOne}
	tournaments, totalTournamentCount, err := s.rdbOperations.GetTournamentList(logger, ctx, tournamentReq)

	if err != nil {
		return entities.GetTournamentTeamVsTeamTableResponse{}, err
	}

	if totalTournamentCount == 0 {
		logger.Error().Err(err).Msg("No tournament found")
		return entities.GetTournamentTeamVsTeamTableResponse{}, error_templates.New("No tournament found", errors.New("no tournament found"), codes.NotFound, http.StatusNotFound)
	}

	var data []entities.DataTournament

	for _, tournament := range tournaments { // Проходим по турнирам нужного города/сезона
		var dataItem entities.DataTournament
		dataItem.TournamentId = tournament.ID
		dataItem.TournamentTypeId = tournament.TypeID
		dataItem.TournamentName = tournament.Name

		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "teamShortName", Name: "Команда"})

		teams, err := s.rdbOperations.FetchTeamsByTournament(logger, ctx, &tournament.ID, &typeOneVsOne)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		if len(teams) == 0 {
			continue
		}
		noGames, err := s.rdbOperations.TeamsHaveNoGamesTournament(logger, ctx, teams, seasonID)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		if noGames { // если нет игр в турнире
			for _, team := range teams { // голы заполняем нулями
				var bodyItem entities.BodyTournament

				bodyItem.Id = team.ID
				bodyItem.TeamShortName = team.ShortName
				bodyItem.Score = 0
				bodyItem.DifferenceInScore = 0
				bodyItem.GamesPlayed = 0
				bodyItem.GamesToPlay = int64(len(teams) - 1) // сейчас только для bo1
				for _, t := range teams {                    // и отображаем нулевой счет
					var cell entities.TableCellTournament

					cell.GameId = 0
					cell.Score = "0:0"

					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
				}
				dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)
			}
			continue
		}

		for _, team := range teams { // Проходим по командам текущего турнира
			var bodyItem entities.BodyTournament
			bodyItem.TableCell = make(map[string][]entities.TableCellTournament, 0)

			gamesToPlay, err := s.rdbOperations.FetchTournamentTeamGames(logger, ctx, tournament.ID, team.ID)
			if err != nil {
				return entities.GetTournamentTeamVsTeamTableResponse{}, err
			}

			bodyItem.Id = team.ID
			bodyItem.TeamName = team.Name
			bodyItem.TeamShortName = team.ShortName
			bodyItem.Score = 0
			bodyItem.DifferenceInScore = 0
			bodyItem.GamesPlayed = 0
			bodyItem.GamesToPlay = int64(len(gamesToPlay))

			for _, t := range teams { // Проходим по командам-соперникам
				var cell entities.TableCellTournament

				cell.GameId = 0
				cell.Score = "0:0"

				if t.ID == team.ID {
					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
					continue //команда сама с собой не играет
				}

				cell.TeamName = t.Name

				tiebreak := false // tiebreak игры не учитываем

				games, err := s.rdbOperations.FetchTournamentPastGames(logger, ctx, team.ID, t.ID, cityID, tournament.ID, &tiebreak)
				if err != nil {
					return entities.GetTournamentTeamVsTeamTableResponse{}, err
				}

				if len(games) == 0 { //если нет игр
					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
					continue
				}

				for _, g := range games {
					cell.GameId = g.ID

					var match1Team1Score int64 = 0
					var match1Team2Score int64 = 0

					if games[0].TechLooseTeamID != nil && *games[0].TechLooseTeamID == team.ID { // Если команда с тех.проигрышем(team.ID) то "30:42" (эта команда проиграла)
						cell.Score = "30:42"
						bodyItem.DifferenceInScore -= 12 // Разницу учитываем
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else if games[0].TechLooseTeamID != nil && *games[0].TechLooseTeamID == t.ID { // Если противника команда с тех.проигрышем(t.ID) то "42:30" (противник проиграл)
						cell.Score = "42:30"
						bodyItem.DifferenceInScore += 12 // Разницу учитываем
						bodyItem.Score += 2              // Добавим себе 2 очка(в целом по игре), если команда противника с тех.проигрышем
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else {
						gamesMatches, err := s.rdbOperations.FetchMatches(logger, ctx, cell.GameId)
						if err != nil {
							return entities.GetTournamentTeamVsTeamTableResponse{}, err
						}

						for _, match := range gamesMatches { // Проходим по матчам
							if match.Team1ID == team.ID {
								match1Team1Score += match.ScoreTeam1
								match1Team2Score += match.ScoreTeam2
							} else {
								match1Team1Score += match.ScoreTeam2
								match1Team2Score += match.ScoreTeam1
							}
						}
						cell.Score = strconv.FormatInt(match1Team1Score, 10) + ":" + strconv.FormatInt(match1Team2Score, 10)

						if len(gamesMatches) != 0 {
							bodyItem.GamesPlayed += 1
							bodyItem.GamesToPlay -= 1
						}
					}

					bodyItem.Score = resumScore(bodyItem.Score, match1Team1Score, match1Team2Score)

					extraPoints, err := s.rdbOperations.GetTournamentTeamExtraPointsCount(logger, ctx, team.ID, tournament.ID)
					if err != nil {
						return entities.GetTournamentTeamVsTeamTableResponse{}, err
					}
					bodyItem.Score += extraPoints // Один раз за турнир считаем дополнительные очки команды

					bodyItem.DifferenceInScore += (match1Team1Score - match1Team2Score)
					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell) // Выставили ячейку со счетом
				}
			} // команды соперников

			dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)

		} // команды текущего турнира
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "score", Name: "Очки"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "differenceInScore", Name: "+/-"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "gamesPlayed", Name: "Игры"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "gamesToPlay", Name: "Осталось"})

		// добавили игры tiebreak текущего турнира
		gamesTiebreak, err := s.rdbOperations.FetchTournamentPastGamesTiebreak(logger, ctx, tournament.ID)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		dataItem.GamesTiebreak = gamesTiebreak

		data = append(data, dataItem)

	} // турниры нужного города

	var response entities.GetTournamentTeamVsTeamTableResponse
	response.DataTournament = data
	response.Message = "OK"
	return response, nil
}

func (s *Service) GetTournamentsTeamVsTeamTable(ctx context.Context, cityID, seasonID int64) (entities.GetTournamentTeamVsTeamTableResponse, error) {
	logger := s.logger.With().Interface("service", "GetTournamentsTeamVsTeamTable").Logger()
	tournamentReq := entities.GetTournamentListRequest{CityID: &cityID, SeasonID: &seasonID, TournamentID: nil, TournamentTypeID: nil}
	tournaments, totalTournamentCount, err := s.rdbOperations.GetTournamentList(logger, ctx, tournamentReq)
	if err != nil {
		return entities.GetTournamentTeamVsTeamTableResponse{}, err
	}

	if totalTournamentCount == 0 {
		logger.Error().Err(err).Msg("No tournament found")
		return entities.GetTournamentTeamVsTeamTableResponse{}, error_templates.New("No tournament found", errors.New("no tournament found"), codes.NotFound, http.StatusNotFound)
	}

	var data []entities.DataTournament

	for _, tournament := range tournaments {
		typeID := tournament.TypeID
		var dataItem entities.DataTournament
		dataItem.TournamentId = tournament.ID
		dataItem.TournamentTypeId = tournament.TypeID
		dataItem.TournamentName = tournament.Name

		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "teamShortName", Name: "Команда"})

		teams, err := s.rdbOperations.FetchTeamsByTournament(logger, ctx, &tournament.ID, &typeID)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		if len(teams) == 0 {
			continue
		}
		teamNames := make(map[int64]string, len(teams))
		for _, tm := range teams {
			teamNames[tm.ID] = tm.Name
		}

		noGames, err := s.rdbOperations.TeamsHaveNoGamesTournament(logger, ctx, teams, seasonID)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		if noGames {
			for _, team := range teams {
				var bodyItem entities.BodyTournament

				bodyItem.Id = team.ID
				bodyItem.TeamShortName = team.ShortName
				bodyItem.Score = 0
				bodyItem.DifferenceInScore = 0
				bodyItem.GamesPlayed = 0
				bodyItem.GamesToPlay = int64(len(teams) - 1) // @TODO: для playoff невозможно посчитать кол-во предстоящих игр.
				// @TODO: скорее всего, нужно оставить поле только для регулярки
				for _, t := range teams {
					var cell entities.TableCellTournament

					cell.GameId = 0
					cell.Score = "0:0"

					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
				}
				dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)
			}
			pu, pd, err := s.fillPlayoffStages(ctx, logger, tournament, teamNames)
			if err != nil {
				return entities.GetTournamentTeamVsTeamTableResponse{}, err
			}
			dataItem.PlayoffUp = pu
			dataItem.PlayoffDown = pd
			gamesTiebreak, err := s.rdbOperations.FetchTournamentPastGamesTiebreak(logger, ctx, tournament.ID)
			if err != nil {
				return entities.GetTournamentTeamVsTeamTableResponse{}, err
			}
			dataItem.GamesTiebreak = gamesTiebreak
			data = append(data, dataItem)
			continue
		}

		for _, team := range teams {
			var bodyItem entities.BodyTournament
			bodyItem.TableCell = make(map[string][]entities.TableCellTournament, 0)

			gamesToPlay, err := s.rdbOperations.FetchTournamentTeamGames(logger, ctx, tournament.ID, team.ID)
			if err != nil {
				return entities.GetTournamentTeamVsTeamTableResponse{}, err
			}

			bodyItem.Id = team.ID
			bodyItem.TeamName = team.Name
			bodyItem.TeamShortName = team.ShortName
			bodyItem.Score = 0
			bodyItem.DifferenceInScore = 0
			bodyItem.GamesPlayed = 0
			bodyItem.GamesToPlay = int64(len(gamesToPlay))

			for _, t := range teams {
				var cell entities.TableCellTournament

				cell.GameId = 0
				cell.Score = "0:0"

				if t.ID == team.ID {
					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
					continue
				}

				cell.TeamName = t.Name

				tiebreak := false

				games, err := s.rdbOperations.FetchTournamentPastGames(logger, ctx, team.ID, t.ID, cityID, tournament.ID, &tiebreak)
				if err != nil {
					return entities.GetTournamentTeamVsTeamTableResponse{}, err
				}

				if len(games) == 0 {
					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
					continue
				}

				for _, g := range games {
					cell.GameId = g.ID

					var match1Team1Score int64 = 0
					var match1Team2Score int64 = 0

					if games[0].TechLooseTeamID != nil && *games[0].TechLooseTeamID == team.ID {
						cell.Score = "30:42"
						bodyItem.DifferenceInScore -= 12
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else if games[0].TechLooseTeamID != nil && *games[0].TechLooseTeamID == t.ID {
						cell.Score = "42:30"
						bodyItem.DifferenceInScore += 12
						bodyItem.Score += 2
						bodyItem.GamesPlayed += 1
						bodyItem.GamesToPlay -= 1
					} else {
						gamesMatches, err := s.rdbOperations.FetchMatches(logger, ctx, cell.GameId)
						if err != nil {
							return entities.GetTournamentTeamVsTeamTableResponse{}, err
						}

						for _, match := range gamesMatches {
							if match.Team1ID == team.ID {
								match1Team1Score += match.ScoreTeam1
								match1Team2Score += match.ScoreTeam2
							} else {
								match1Team1Score += match.ScoreTeam2
								match1Team2Score += match.ScoreTeam1
							}
						}
						cell.Score = strconv.FormatInt(match1Team1Score, 10) + ":" + strconv.FormatInt(match1Team2Score, 10)

						if len(gamesMatches) != 0 {
							bodyItem.GamesPlayed += 1
							bodyItem.GamesToPlay -= 1
						}
					}

					bodyItem.Score = resumScore(bodyItem.Score, match1Team1Score, match1Team2Score)

					extraPoints, err := s.rdbOperations.GetTournamentTeamExtraPointsCount(logger, ctx, team.ID, tournament.ID)
					if err != nil {
						return entities.GetTournamentTeamVsTeamTableResponse{}, err
					}
					bodyItem.Score += extraPoints

					bodyItem.DifferenceInScore += (match1Team1Score - match1Team2Score)
					bodyItem.TableCell[t.ShortName] = append(bodyItem.TableCell[t.ShortName], cell)
				}
			}

			dataItem.Table.Body = append(dataItem.Table.Body, bodyItem)

		}
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "score", Name: "Очки"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "differenceInScore", Name: "+/-"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "gamesPlayed", Name: "Игры"})
		dataItem.Table.Columns = append(dataItem.Table.Columns, entities.Column{Uid: "gamesToPlay", Name: "Осталось"})

		gamesTiebreak, err := s.rdbOperations.FetchTournamentPastGamesTiebreak(logger, ctx, tournament.ID)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		dataItem.GamesTiebreak = gamesTiebreak

		pu, pd, err := s.fillPlayoffStages(ctx, logger, tournament, teamNames)
		if err != nil {
			return entities.GetTournamentTeamVsTeamTableResponse{}, err
		}
		dataItem.PlayoffUp = pu
		dataItem.PlayoffDown = pd

		data = append(data, dataItem)
	}

	var response entities.GetTournamentTeamVsTeamTableResponse
	response.DataTournament = data
	response.Message = "OK"
	return response, nil
}

func parseTournamentStageOrder(number string) (int64, error) {
	parts := strings.Split(number, constant.SeparatorStageNumber)
	if len(parts) == 0 || parts[0] == "" {
		return 0, errors.New("empty stage number")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func bracketTeamSets(b entities.TournamentBrackets) (wb, lb map[int64]struct{}) {
	wb = make(map[int64]struct{})
	lb = make(map[int64]struct{})
	for _, p := range b.TeamsOfWinnerBracket {
		if p != nil {
			wb[*p] = struct{}{}
		}
	}
	for _, p := range b.TeamsOfLoserBracket {
		if p != nil {
			lb[*p] = struct{}{}
		}
	}
	return wb, lb
}

func (s *Service) tournamentGameScoreString(logger zerolog.Logger, ctx context.Context, g entities.TournamentGame) (string, error) {
	if g.TechLooseTeamID != nil {
		if *g.TechLooseTeamID == g.Team1ID {
			return "30:42", nil
		}
		if *g.TechLooseTeamID == g.Team2ID {
			return "42:30", nil
		}
	}
	matches, err := s.rdbOperations.FetchMatches(logger, ctx, g.ID)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "0:0", nil
	}
	var s1, s2 int64
	for _, m := range matches {
		if m.Team1ID == g.Team1ID {
			s1 += m.ScoreTeam1
			s2 += m.ScoreTeam2
		} else {
			s1 += m.ScoreTeam2
			s2 += m.ScoreTeam1
		}
	}
	return strconv.FormatInt(s1, 10) + ":" + strconv.FormatInt(s2, 10), nil
}

func (s *Service) tournamentGameInfo(ctx context.Context, logger zerolog.Logger, g entities.TournamentGame, teamNames map[int64]string) (entities.GameInfo, error) {
	score, err := s.tournamentGameScoreString(logger, ctx, g)
	if err != nil {
		return entities.GameInfo{}, err
	}
	t1 := teamNames[g.Team1ID]
	t2 := teamNames[g.Team2ID]
	if t1 == "" {
		logger.Error().Err(err).Msg("failed to s.tournamentGameInfo: could not get a teamName by teamId")
		return entities.GameInfo{}, errors.New("failed to s.tournamentGameInfo: could not get teamName by teamId")
	}
	if t2 == "" {
		logger.Error().Err(err).Msg("failed to s.tournamentGameInfo: could not get a teamName by teamId")
		return entities.GameInfo{}, errors.New("failed to s.tournamentGameInfo: could not get teamName by teamId")
	}
	return entities.GameInfo{GameId: g.ID, Score: score, Team1Name: t1, Team2Name: t2}, nil
}

func (s *Service) fillPlayoffStages(ctx context.Context, logger zerolog.Logger, tournament entities.TournamentShort, teamNames map[int64]string) (playoffUp, playoffDown []entities.Stages, err error) {
	switch tournament.TypeID {
	case constant.RegularTournamentTypeID, constant.RegularOneVsOneTournamentTypeID:
		return nil, nil, nil
	}

	stages, err := s.rdbOperations.GetTournamentStageList(ctx, logger, tournament.ID)
	if err != nil {
		return nil, nil, err
	}

	slices.SortFunc(stages, func(a, b entities.TournamentStage) int {
		oa, errA := parseTournamentStageOrder(a.Number)
		ob, errB := parseTournamentStageOrder(b.Number)
		if errA != nil || errB != nil {
			// если не удалось распарсить номер этапа, то сортируем по id стадии
			return int(a.ID - b.ID)
		}
		if oa < ob {
			return -1
		}
		if oa > ob {
			return 1
		}
		// в случае, если каким-то образом номера стадий совпали, сортируем по id стадии
		return int(a.ID - b.ID)
	})

	for _, st := range stages {
		orderNum, err := parseTournamentStageOrder(st.Number)
		if err != nil {
			continue
		}

		games, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, st.ID)
		if err != nil {
			return nil, nil, err
		}
		sort.Slice(games, func(i, j int) bool { return games[i].ID < games[j].ID })

		if tournament.TypeID == constant.RegularPlayoffWithLoserBracketTournamentTypeID {
			bracket, err := s.rdbOperations.GetTeamsFromTournamentBrackets(ctx, logger, tournament.ID, orderNum)
			if err != nil {
				return nil, nil, err
			}
			wbSet, lbSet := bracketTeamSets(bracket)
			var upGames, downGames []entities.GameInfo

			// Если оба множества пустые, то данных в brackets нет.
			// Тогда все игры этапа попадают только в upGames — дальше они уйдут в PlayoffUp.
			if len(wbSet) == 0 && len(lbSet) == 0 {
				for _, g := range games {
					gi, err := s.tournamentGameInfo(ctx, logger, g, teamNames)
					if err != nil {
						return nil, nil, err
					}
					upGames = append(upGames, gi)
				}
			} else {
				for _, g := range games {
					gi, err := s.tournamentGameInfo(ctx, logger, g, teamNames)
					if err != nil {
						return nil, nil, err
					}
					// Проверяем, к какой сетке принадлежит команда на данной стадии
					_, inWb1 := wbSet[g.Team1ID]
					_, inWb2 := wbSet[g.Team2ID]
					_, inLb1 := lbSet[g.Team1ID]
					_, inLb2 := lbSet[g.Team2ID]
					switch {
					case inWb1 && inWb2:
						upGames = append(upGames, gi)
					case inLb1 && inLb2:
						downGames = append(downGames, gi)
					// любая другая комбинация (кросс-сеточные пары, финал и т.д.)
					default:
						upGames = append(upGames, gi)
					}
				}
			}
			if len(upGames) > 0 {
				sort.Slice(upGames, func(i, j int) bool { return upGames[i].GameId < upGames[j].GameId })
				playoffUp = append(playoffUp, entities.Stages{StageId: st.ID, StageNumber: st.Number, Games: upGames})
			}
			if len(downGames) > 0 {
				sort.Slice(downGames, func(i, j int) bool { return downGames[i].GameId < downGames[j].GameId })
				playoffDown = append(playoffDown, entities.Stages{StageId: st.ID, StageNumber: st.Number, Games: downGames})
			}
			continue
		}

		var allGames []entities.GameInfo
		for _, g := range games {
			gi, err := s.tournamentGameInfo(ctx, logger, g, teamNames)
			if err != nil {
				return nil, nil, err
			}
			allGames = append(allGames, gi)
		}
		sort.Slice(allGames, func(i, j int) bool { return allGames[i].GameId < allGames[j].GameId })
		playoffUp = append(playoffUp, entities.Stages{StageId: st.ID, StageNumber: st.Number, Games: allGames})
	}
	return playoffUp, playoffDown, nil
}

func (s *Service) Create(ctx context.Context, request *entities.CreateTeamRequest) (int64, error) {
	logger := s.logger.With().Str("service", "Create").Logger()

	if request.Creator.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID)
		if err != nil {
			return 0, err
		}

		if request.CityIdParam == "" {

			request.CityId = master.City.ID

		} else {

			cityId, err := strconv.ParseInt(request.CityIdParam, 10, 64)
			if err != nil {
				err = errors.New(pkgerr.WrongParameterError + ": " + "cityId")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			if cityId != master.City.ID {
				err = errors.New(pkgerr.ErrCityIdNotEqualMasterCityId)
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			request.CityId = cityId
		}

	} else {

		if request.CityIdParam == "" {
			err := errors.New(pkgerr.EmptyParameterError + ": " + "cityId")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		cityId, err := strconv.ParseInt(request.CityIdParam, 10, 64)
		if err != nil {
			err = errors.New(pkgerr.WrongParameterError + ": " + "cityId")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		request.CityId = cityId
	}

	id, err := s.rwdbOperations.CreateTeam(logger, ctx, *request, nil)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, request *entities.UpdateTeamRequest) (bool, error) {
	logger := s.logger.With().Str("service", "Update").Logger()

	if request.Updater.Role != nil && request.Updater.Role.Name == constant.TournamentMaster {
		team, err := s.rdbOperations.GetTeamById(logger, ctx, request.ID, nil)
		if err != nil {
			return false, err
		}

		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Updater.ID)
		if err != nil {
			return false, err
		}

		if team.CityId == nil || *team.CityId != master.City.ID {
			err = fmt.Errorf(
				"город команды(ID: %s) не совпадает с городом мастера по турнирам(ID: %d)",
				convert.IntPtrToStr[int64](team.CityId),
				master.City.ID,
			)
			return false, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if request.CityId != nil && *request.CityId != master.City.ID {
			err = fmt.Errorf(
				"мастер по турнирам может работать только с командами своего города(ID: %d), ID города в запросе - %d",
				master.City.ID,
				*request.CityId,
			)
			logger.Error().Err(err).Msg("failed to team.Update")
			return false, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	res, err := s.rwdbOperations.UpdateTeam(logger, ctx, *request, &s.config.RWDB)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (bool, error) {
	logger := s.logger.With().Interface("service", "Delete").Logger()

	res, err := s.rwdbOperations.DeleteTeam(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) AddPlayerIntoTeam(ctx context.Context, req *entities.MovingPlayerTeam) (bool, error) {
	logger := s.logger.With().Str("service", "AddPlayerIntoTeam").Logger()

	player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(req.PlayerID), nil)
	if err != nil {
		return false, err
	}

	team, err := s.rdbOperations.GetTeamById(logger, ctx, req.TeamID, nil)
	if err != nil {
		return false, err
	}

	if int64(pointer.GetValue(player.CityID)) != pointer.GetValue(team.CityId) {
		err = fmt.Errorf("город(id:%d) игрока %s не равен городу(id:%d) команды %s", pointer.GetValue(player.CityID), pointer.GetValue(player.Name), pointer.GetValue(team.CityId), team.Name)
		return false, err
	}

	if req.Executor.Role != nil && req.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
		if err != nil {
			return false, err
		}

		if player.CityID == nil || int64(*player.CityID) != master.City.ID {
			err = fmt.Errorf(
				"город игрока(ID: %s) не совпадает с городом мастера по турнирам(ID: %d)",
				convert.IntPtrToStr[int](player.CityID),
				master.City.ID,
			)
			logger.Error().Err(err).Msg("failed team.AddPlayerIntoTeam")
			return false, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if team.CityId == nil || *team.CityId != master.City.ID {
			err = fmt.Errorf(
				"город команды(ID: %s) не совпадает с городом мастера по турнирам(ID: %d)",
				convert.IntPtrToStr[int64](team.CityId),
				master.City.ID,
			)
			logger.Error().Err(err).Msg("failed team.AddPlayerIntoTeam")
			return false, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	res, err := s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, req.PlayerID, req.TeamID, nil)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s *Service) RemovePlayerFromTeam(ctx context.Context, req *entities.MovingPlayerTeam) (bool, error) {
	logger := s.logger.With().Str("service", "RemovePlayerFromTeam").Logger()

	if req.Executor.Role != nil && req.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
		if err != nil {
			return false, err
		}

		player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(req.PlayerID), nil)
		if err != nil {
			return false, err
		}

		team, err := s.rdbOperations.GetTeamById(logger, ctx, req.TeamID, nil)
		if err != nil {
			return false, err
		}

		if player.CityID == nil || int64(*player.CityID) != master.City.ID {
			err = fmt.Errorf(
				"город игрока(ID: %s) не совпадает с городом мастера по турнирам(ID: %d)",
				convert.IntPtrToStr[int](player.CityID),
				master.City.ID,
			)
			logger.Error().Err(err).Msg("failed team.AddPlayerIntoTeam")
			return false, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if team.CityId == nil || *team.CityId != master.City.ID {
			err = fmt.Errorf(
				"город команды(ID: %s) не совпадает с городом мастера по турнирам(ID: %d)",
				convert.IntPtrToStr[int64](team.CityId),
				master.City.ID,
			)
			logger.Error().Err(err).Msg("failed team.AddPlayerIntoTeam")
			return false, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	res, err := s.rwdbOperations.RemovePlayerFromTeam(logger, ctx, req.PlayerID, req.TeamID, &s.config.RWDB)
	if err != nil {
		return res, err
	}

	return res, nil
}

func findBestPlayer(players []entities.FullPlayer, leagueId int) entities.FullPlayer {
	var bestPlayer = entities.FullPlayer{}

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
