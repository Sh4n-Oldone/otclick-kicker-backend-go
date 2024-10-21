package helpers

import "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"

func ConvertCreateCityRequestToCity(request *entity.CreateCityRequest) *entity.City{
	return &entity.City{
		Name: request.Name,
		Ru: request.Ru,
		Deleted: request.Deleted,
	}
}

func ConvertUpdateCityRequestToCity(request *entity.UpdateCityRequest) *entity.City{
	return &entity.City{
		ID: request.ID,
		Name: request.Name,
		Ru: request.Ru,
		Deleted: request.Deleted,
	}
}