package errors

const (
	FailedValidateRequest = "ошибка валидации запроса"
	FailedCastRequest     = "ошибка преобразования запроса"
	EmptyParameterError   = "отсутствует параметр в запросе"
	WrongParameterError   = "неверный параметр в запросе"
	ErrEmptyLastName      = "отсутствует фамилия игрока"
)

// Database errors

const (
	ErrCreatePlayer      = "ошибка создания игрока"
	ErrDeletePlayer      = "ошибка удаления игрока"
	ErrUpdatePlayer      = "ошибка обновления данных игрока"
	ErrGetPlayer         = "ошибка получения данных игрока"
	ErrGetPlayerList     = "ошибка получения списка игроков"
	ErrGetGame           = "ошибка получения данных игры"
	ErrGetGameList       = "ошибка получения списка игр"
	ErrGetMatches        = "ошибка получения данных матчей"
	ErrGetLeague         = "ошибка получения данных лиги"
	ErrGetLeagueList     = "ошибка получения списка лиг"
	ErrPlayerDontUpdated = "данные игрока не были обновлены"
	ErrPlayerDontDeleted = "игрок не был удален"
)
