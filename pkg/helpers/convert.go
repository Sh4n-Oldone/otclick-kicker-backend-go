package helpers

import "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"

func ConvertCreateCityRequestToCity(request *entity.CreateCityRequest) *entity.City {
	return &entity.City{
		Name:    request.Name,
		Ru:      request.Ru,
		DeletedAt: request.DeletedAt,
	}
}

func ConvertUpdateCityRequestToCity(request *entity.UpdateCityRequest) *entity.City {
	return &entity.City{
		ID:      request.ID,
		Name:    request.Name,
		Ru:      request.Ru,
		DeletedAt: request.DeletedAt,
	}
}

func ConvertCreateUserRequestToUser(request *entity.CreateUserRequest) *entity.User{
	return &entity.User{
		Email: request.Email,
		Password: []byte(request.Password),
	}
}

func ConvertLoginUserRequestToUser(request *entity.LoginUserRequest) *entity.User{
	return &entity.User{
		Email: request.Email,
		Password: []byte(request.Password),
	}
}

func ConvertCreateMatchRequestToMatch(request *entity.CreateMatchRequest) *entity.Match {
	return &entity.Match{
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

func ConvertUpdateMatchRequestToMatch(request *entity.UpdateMatchRequest) *entity.Match {
	return &entity.Match{
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

