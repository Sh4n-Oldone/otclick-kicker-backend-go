package errors

const (
	FailedValidateRequest = "ошибка валидации запроса"
	FailedCastRequest     = "ошибка преобразования запроса"
	EmptyParameterError   = "отсутствует параметр в запросе"
	WrongParameterError   = "неверный параметр в запросе"
	ErrEmptyLastName      = "отсутствует фамилия игрока"
	ValidationErr         = "ошибка валидации"
	ErrNoMatches          = "отсутствует список матчей в запросе"
	ErrNoTeam             = "отсутствует id команды в запросе"
	ErrNoDate             = "отсутствует дата игры в запросе"
	ErrWrongDate          = "некорректная дата игры в запросе"
)

// Database errors

const (
	ErrCreatePlayer        = "ошибка создания игрока"
	ErrDeletePlayer        = "ошибка удаления игрока"
	ErrUpdatePlayer        = "ошибка обновления данных игрока"
	ErrGetPlayer           = "ошибка получения данных игрока"
	ErrGetPlayerList       = "ошибка получения списка игроков"
	ErrGetGame             = "ошибка получения данных игры"
	ErrGetGameList         = "ошибка получения списка игр"
	ErrGetMatches          = "ошибка получения данных матчей"
	ErrGetLeague           = "ошибка получения данных лиги"
	ErrGetLeagueList       = "ошибка получения списка лиг"
	ErrTeamAlreadyInLeague = "команда уже находится в составе другой лиги"
	ErrCreateGame          = "ошибка создания игры"
	ErrPlayerNotFound      = "игрок не найден"
	ErrCreateMatch         = "ошибка создания матча"
	ErrDeleteGame          = "ошибка удаления игры"
	ErrGameNotFound        = "игра не найдена"
	ErrDeleteMatch         = "ошибка удаления матча"
	ErrUpdateGame          = "ошибка обновления данных игры"
	ErrUpdateMatch         = "ошибка обновления данных матча"
)
