package errors

const (
	FailedValidateRequest              = "ошибка валидации запроса"
	FailedCastRequest                  = "ошибка преобразования запроса"
	EmptyParameterError                = "отсутствует параметр в запросе"
	WrongParameterError                = "неверное значение параметра в запросе"
	ShortParameterError                = "слишком короткое значение параметра в запросе"
	ErrDifferentTeams                  = "команды в матчах и в игре не совпадают "
	ErrWrongDate                       = "некорректная дата игры в запросе"
	WrongUserRole                      = "некорректная роль пользователя"
	FailedParseJWTToken                = "ошибка расшифровки JWT токена"
	FailedGetUserData                  = "ошибка получения данных пользователя"
	FailedGameByTeamsLeagueMismatch    = "ошибка совмещения команд из разных лиг"
	ErrEmptyField                      = "отсутствует значение в поле запроса"
	ErrUserIdToken                     = "не удалось получить идентификатор пользователя из токена"
	ErrRoleToken                       = "не удалось получить роли из токена"
	ErrCityIdNotEqualMasterCityId      = "cityId должен быть равен cityId мастера по турнирам"
	ErrBarNotInMasterCity              = "бар находится не в городе мастера по турнирам"
	ErrCityLeagueAndMasterMismatch     = "город лиги и город мастера по турнирам не совпадают"
	ErrCityTournamentAndMasterMismatch = "город турнира и город мастера по турнирам не совпадают"
	ErrGameIsNotPartOfLeague           = "игра была провдена не в рамках лиги"
	ErrDeleteTeamFromLeague            = "ошибка удаления команды из лиги, у команды %d есть прошлые игры в лиге %d"
)

// Database errors
const (
	ErrPlayerNotFound = "игрок не найден"
	ErrGameNotFound   = "игра не найдена"
)
