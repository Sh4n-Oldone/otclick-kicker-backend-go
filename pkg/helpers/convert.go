package helpers

import "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"

func ConvertCreateCityRequestToCity(request *entities.CreateCityRequest) *entities.City {
	return &entities.City{
		Name:      request.Name,
		Ru:        request.Ru,
		DeletedAt: request.DeletedAt,
	}
}

func ConvertUpdateCityRequestToCity(request *entities.UpdateCityRequest) *entities.City {
	return &entities.City{
		ID:   request.ID,
		Name: request.Name,
		Ru:   request.Ru,
	}
}

func ConvertCreateUserRequestToUser(request *entities.CreateUserRequest) *entities.User {
	return &entities.User{
		Email:    request.Email,
		Password: []byte(request.Password),
	}
}

func ConvertLoginUserRequestToUser(request *entities.LoginUserRequest) *entities.User {
	return &entities.User{
		Email:    request.Email,
		Password: []byte(request.Password),
	}
}

func ConvertCreateMatchRequestToMatch(request *entities.CreateMatchRequest) *entities.Match {
	return &entities.Match{
		Date:           request.Date,
		GameID:         request.GameID,
		Team1ID:        request.Team1ID,
		Team2ID:        request.Team2ID,
		Player1Team1ID: request.Player1Team1ID,
		Player2Team1ID: request.Player2Team1ID,
		Player1Team2ID: request.Player1Team2ID,
		Player2Team2ID: request.Player2Team2ID,
		ScoreTeam1:     request.ScoreTeam1,
		ScoreTeam2:     request.ScoreTeam2,
	}
}

func ConvertUpdateMatchRequestToMatch(request *entities.UpdateMatchRequest) *entities.Match {
	return &entities.Match{
		ID:             request.ID,
		Date:           request.Date,
		GameID:         request.GameID,
		Team1ID:        request.Team1ID,
		Team2ID:        request.Team2ID,
		Player1Team1ID: request.Player1Team1ID,
		Player2Team1ID: request.Player2Team1ID,
		Player1Team2ID: request.Player1Team2ID,
		Player2Team2ID: request.Player2Team2ID,
		ScoreTeam1:     request.ScoreTeam1,
		ScoreTeam2:     request.ScoreTeam2,
	}
}

func ConvertCreateLeagueRequestToLeague(request *entities.CreateLeagueRequest) *entities.League {
	return &entities.League{
		CityID:   request.CityID,
		Name:     request.Name,
		SeasonID: request.SeasonID,
	}
}

func ConvertUpdateLeagueRequestToLeague(request *entities.UpdateLeagueRequest) *entities.League {
	return &entities.League{
		ID:       request.ID,
		Name:     request.Name,
		SeasonID: request.SeasonID,
	}
}
