package endpoint

import (
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/city"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/match"
)

type ServicesEndpoints struct {
	CityEP city.Endpoints
	MatchEP match.Endpoints
}
