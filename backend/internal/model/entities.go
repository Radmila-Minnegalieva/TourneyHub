// Package model содержит значения и параметры предметной области без зависимости от HTTP и PostgreSQL.
package model

import (
	"github.com/google/uuid"
	"time"
)

// Допустимые значения ApplicationStatus.
const (
	ApplicationStatusAccepted  ApplicationStatus = "accepted"
	ApplicationStatusPending   ApplicationStatus = "pending"
	ApplicationStatusRejected  ApplicationStatus = "rejected"
	ApplicationStatusWithdrawn ApplicationStatus = "withdrawn"
)

// Допустимые значения DisputeStatus.
const (
	DisputeStatusOpen     DisputeStatus = "open"
	DisputeStatusResolved DisputeStatus = "resolved"
)

// Допустимые значения FileContentType.
const (
	Applicationpdf FileContentType = "application/pdf"
	Imagejpeg      FileContentType = "image/jpeg"
	Imagepng       FileContentType = "image/png"
)

// Допустимые значения FilePurpose.
const (
	FilePurposeAvatar        FilePurpose = "avatar"
	FilePurposeMatchProtocol FilePurpose = "match_protocol"
	FilePurposeTeamLogo      FilePurpose = "team_logo"
)

// Valid проверяет, входит ли значение в перечисление FilePurpose.
func (e FilePurpose) Valid() bool {
	switch e {
	case FilePurposeAvatar:
		return true
	case FilePurposeMatchProtocol:
		return true
	case FilePurposeTeamLogo:
		return true
	default:
		return false
	}
}

// Допустимые значения InvitationStatus.
const (
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

// Допустимые значения InvitationDecisionRequestDecision.
const (
	InvitationDecisionRequestDecisionAccept  InvitationDecisionRequestDecision = "accept"
	InvitationDecisionRequestDecisionDecline InvitationDecisionRequestDecision = "decline"
)

// Допустимые значения MatchLoserNextMatchSlot.
const (
	MatchLoserNextMatchSlotAway MatchLoserNextMatchSlot = "away"
	MatchLoserNextMatchSlotHome MatchLoserNextMatchSlot = "home"
)

// Допустимые значения MatchNextMatchSlot.
const (
	MatchNextMatchSlotAway MatchNextMatchSlot = "away"
	MatchNextMatchSlotHome MatchNextMatchSlot = "home"
)

// Допустимые значения MatchStatus.
const (
	MatchStatusAwaitingConfirmation MatchStatus = "awaiting_confirmation"
	MatchStatusBye                  MatchStatus = "bye"
	MatchStatusCompleted            MatchStatus = "completed"
	MatchStatusDisputed             MatchStatus = "disputed"
	MatchStatusScheduled            MatchStatus = "scheduled"
)

// Допустимые значения RegisterRequestPersonalDataConsent.
const (
	True RegisterRequestPersonalDataConsent = true
)

// Допустимые значения ResultReportStatus.
const (
	ResultReportStatusConfirmed  ResultReportStatus = "confirmed"
	ResultReportStatusDisputed   ResultReportStatus = "disputed"
	ResultReportStatusPending    ResultReportStatus = "pending"
	ResultReportStatusSuperseded ResultReportStatus = "superseded"
)

// Допустимые значения ReviewApplicationRequestDecision.
const (
	ReviewApplicationRequestDecisionAccept ReviewApplicationRequestDecision = "accept"
	ReviewApplicationRequestDecisionReject ReviewApplicationRequestDecision = "reject"
)

// Valid проверяет, входит ли значение в перечисление ReviewApplicationRequestDecision.
func (e ReviewApplicationRequestDecision) Valid() bool {
	switch e {
	case ReviewApplicationRequestDecisionAccept:
		return true
	case ReviewApplicationRequestDecisionReject:
		return true
	default:
		return false
	}
}

// Допустимые значения Role.
const (
	Admin       Role = "admin"
	Organizer   Role = "organizer"
	Participant Role = "participant"
	Referee     Role = "referee"
)

// Допустимые значения RulesBestOf.
const (
	RulesBestOfN1 RulesBestOf = 1
	RulesBestOfN3 RulesBestOf = 3
	RulesBestOfN5 RulesBestOf = 5
)

// Valid проверяет, входит ли значение в перечисление RulesBestOf.
func (e RulesBestOf) Valid() bool {
	switch e {
	case RulesBestOfN1:
		return true
	case RulesBestOfN3:
		return true
	case RulesBestOfN5:
		return true
	default:
		return false
	}
}

// Допустимые значения RulesRoundRobinLegs.
const (
	RulesRoundRobinLegsN1 RulesRoundRobinLegs = 1
	RulesRoundRobinLegsN2 RulesRoundRobinLegs = 2
)

// Valid проверяет, входит ли значение в перечисление RulesRoundRobinLegs.
func (e RulesRoundRobinLegs) Valid() bool {
	switch e {
	case RulesRoundRobinLegsN1:
		return true
	case RulesRoundRobinLegsN2:
		return true
	default:
		return false
	}
}

// Допустимые значения SeedingRequestMethod.
const (
	SeedingRequestMethodElo    SeedingRequestMethod = "elo"
	SeedingRequestMethodManual SeedingRequestMethod = "manual"
	SeedingRequestMethodRandom SeedingRequestMethod = "random"
)

// Допустимые значения StageStatus.
const (
	StageStatusCompleted  StageStatus = "completed"
	StageStatusInProgress StageStatus = "in_progress"
	StageStatusPending    StageStatus = "pending"
)

// Допустимые значения StageKind.
const (
	StageKindGroups     StageKind = "groups"
	StageKindPlayoffs   StageKind = "playoffs"
	StageKindRoundRobin StageKind = "round_robin"
)

// Допустимые значения TieBreaker.
const (
	TieBreakerDraw            TieBreaker = "draw"
	TieBreakerElo             TieBreaker = "elo"
	TieBreakerHeadToHead      TieBreaker = "head_to_head"
	TieBreakerScoreDifference TieBreaker = "score_difference"
	TieBreakerScoreFor        TieBreaker = "score_for"
)

// Valid проверяет, входит ли значение в перечисление TieBreaker.
func (e TieBreaker) Valid() bool {
	switch e {
	case TieBreakerDraw:
		return true
	case TieBreakerElo:
		return true
	case TieBreakerHeadToHead:
		return true
	case TieBreakerScoreDifference:
		return true
	case TieBreakerScoreFor:
		return true
	default:
		return false
	}
}

// Допустимые значения TokensExpiresIn.
const (
	N900 TokensExpiresIn = 900
)

// Допустимые значения TokensTokenType.
const (
	Bearer TokensTokenType = "Bearer"
)

// Допустимые значения TournamentFormat.
const (
	TournamentFormatGroupsPlayoffs    TournamentFormat = "groups_playoffs"
	TournamentFormatRoundRobin        TournamentFormat = "round_robin"
	TournamentFormatSingleElimination TournamentFormat = "single_elimination"
)

// Valid проверяет, входит ли значение в перечисление TournamentFormat.
func (e TournamentFormat) Valid() bool {
	switch e {
	case TournamentFormatGroupsPlayoffs:
		return true
	case TournamentFormatRoundRobin:
		return true
	case TournamentFormatSingleElimination:
		return true
	default:
		return false
	}
}

// Допустимые значения TournamentStatus.
const (
	TournamentStatusCancelled    TournamentStatus = "cancelled"
	TournamentStatusCompleted    TournamentStatus = "completed"
	TournamentStatusDraft        TournamentStatus = "draft"
	TournamentStatusInProgress   TournamentStatus = "in_progress"
	TournamentStatusRegistration TournamentStatus = "registration"
)

// Допустимые значения ListAssignedDisputesParamsStatus.
const (
	ListAssignedDisputesParamsStatusOpen     ListAssignedDisputesParamsStatus = "open"
	ListAssignedDisputesParamsStatusResolved ListAssignedDisputesParamsStatus = "resolved"
)

// Application описывает модель Application.
type Application struct {
	CreatedAt       time.Time         `json:"created_at"`
	FinalPlace      *int              `json:"final_place,omitempty"`
	Id              uuid.UUID         `json:"id"`
	Players         []PublicUser      `json:"players"`
	RejectionReason *string           `json:"rejection_reason,omitempty"`
	Seed            *int              `json:"seed,omitempty"`
	Status          ApplicationStatus `json:"status"`
	Team            Team              `json:"team"`
	TournamentId    uuid.UUID         `json:"tournament_id"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// ApplicationPage описывает модель ApplicationPage.
type ApplicationPage struct {
	Items  []Application `json:"items"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
	Total  int           `json:"total"`
}

// ApplicationRequest Капитан подаёт за team_id; для индивидуальной дисциплины team_id можно опустить и player_ids содержит только текущего пользователя: персональная команда создаётся автоматически. Состав фиксируется снимком; один игрок занят только одной pending/accepted заявкой турнира.
type ApplicationRequest struct {
	PlayerIds []uuid.UUID `json:"player_ids"`
	TeamId    *uuid.UUID  `json:"team_id,omitempty"`
}

// ApplicationStatus описывает модель ApplicationStatus.
type ApplicationStatus string

// AssignRefereeRequest описывает модель AssignRefereeRequest.
type AssignRefereeRequest struct {
	RefereeId uuid.UUID `json:"referee_id"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// AuditEntry Не содержит паролей, токенов и секретов. Результаты, посев, роли и модерация записываются в транзакции с изменением данных.
type AuditEntry struct {
	Action     string                  `json:"action"`
	ActorId    *uuid.UUID              `json:"actor_id,omitempty"`
	After      *map[string]interface{} `json:"after,omitempty"`
	Before     *map[string]interface{} `json:"before,omitempty"`
	CreatedAt  time.Time               `json:"created_at"`
	EntityId   uuid.UUID               `json:"entity_id"`
	EntityType string                  `json:"entity_type"`
	Id         uuid.UUID               `json:"id"`
}

// AuditEntryPage описывает модель AuditEntryPage.
type AuditEntryPage struct {
	Items  []AuditEntry `json:"items"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

// BlockUserRequest При блокировке причина обязательна; активные сессии отзываются. Последний администратор защищён.
type BlockUserRequest struct {
	Blocked bool    `json:"blocked"`
	Reason  *string `json:"reason,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// Bracket описывает модель Bracket.
type Bracket struct {
	Groups  []Group `json:"groups"`
	Matches []Match `json:"matches"`
	Stages  []Stage `json:"stages"`
}

// ConfirmDrawRequest Повторное подтверждение того же preview_id идемпотентно. Пересборка разрешена до первого сообщения/фиксации результата.
type ConfirmDrawRequest struct {
	PreviewId uuid.UUID `json:"preview_id"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// CreateInvitationRequest nickname задан: адресное приглашение; пустой объект: одноразовая ссылка. Срок действия 7 дней.
type CreateInvitationRequest struct {
	Nickname *string `json:"nickname,omitempty"`
}

// CreateTeamRequest описывает модель CreateTeamRequest.
type CreateTeamRequest struct {
	Description *string    `json:"description,omitempty"`
	LogoFileId  *uuid.UUID `json:"logo_file_id,omitempty"`
	Name        string     `json:"name"`
}

// CreateTournamentRequest Черновик может не иметь дат. Публикация требует opens < closes <= starts < ends; timezone — IANA.
type CreateTournamentRequest struct {
	Description          *string          `json:"description,omitempty"`
	DisciplineId         uuid.UUID        `json:"discipline_id"`
	EndsAt               *time.Time       `json:"ends_at,omitempty"`
	Format               TournamentFormat `json:"format"`
	MaxParticipants      int              `json:"max_participants"`
	Name                 string           `json:"name"`
	RegistrationClosesAt *time.Time       `json:"registration_closes_at,omitempty"`
	RegistrationOpensAt  *time.Time       `json:"registration_opens_at,omitempty"`

	// Rules group_settings обязателен только для groups_playoffs; third_place_match применим к плей-офф. Бо-серия: победитель выигрывает (best_of+1)/2 партий; правила ничьей проверяются по дисциплине/этапу.
	Rules    Rules      `json:"rules"`
	StartsAt *time.Time `json:"starts_at,omitempty"`
	Timezone string     `json:"timezone"`
}

// Discipline описывает модель Discipline.
type Discipline struct {
	Active        bool      `json:"active"`
	AllowsDraw    bool      `json:"allows_draw"`
	Id            uuid.UUID `json:"id"`
	MaxRosterSize int       `json:"max_roster_size"`
	MinRosterSize int       `json:"min_roster_size"`
	Name          string    `json:"name"`
	ScoreUnit     string    `json:"score_unit"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// DisciplinePage описывает модель DisciplinePage.
type DisciplinePage struct {
	Items  []Discipline `json:"items"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

// DisciplineRequest min_roster_size <= max_roster_size. Индивидуальная дисциплина: оба равны 1.
type DisciplineRequest struct {
	AllowsDraw    bool   `json:"allows_draw"`
	MaxRosterSize int    `json:"max_roster_size"`
	MinRosterSize int    `json:"min_roster_size"`
	Name          string `json:"name"`
	ScoreUnit     string `json:"score_unit"`
}

// Dispute описывает модель Dispute.
type Dispute struct {
	CreatedAt       time.Time     `json:"created_at"`
	EvidenceFileIds []uuid.UUID   `json:"evidence_file_ids"`
	Id              uuid.UUID     `json:"id"`
	MatchId         uuid.UUID     `json:"match_id"`
	OpenedBy        uuid.UUID     `json:"opened_by"`
	Reason          string        `json:"reason"`
	ReportId        uuid.UUID     `json:"report_id"`
	Resolution      *string       `json:"resolution,omitempty"`
	ResolvedBy      *uuid.UUID    `json:"resolved_by,omitempty"`
	Status          DisputeStatus `json:"status"`
}

// DisputeStatus описывает модель Dispute.Status.
type DisputeStatus string

// DisputePage описывает модель DisputePage.
type DisputePage struct {
	Items  []Dispute `json:"items"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Total  int       `json:"total"`
}

// DisputeRequest описывает модель DisputeRequest.
type DisputeRequest struct {
	EvidenceFileIds *[]uuid.UUID `json:"evidence_file_ids,omitempty"`
	Reason          string       `json:"reason"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// DrawPreview Предпросмотр хранится 1 час; подтверждение использует именно эту раскладку, не повторяет случайную генерацию.
type DrawPreview struct {
	ExpiresAt    time.Time `json:"expires_at"`
	Groups       []Group   `json:"groups"`
	Id           uuid.UUID `json:"id"`
	Matches      []Match   `json:"matches"`
	Seeds        []Seed    `json:"seeds"`
	Stages       []Stage   `json:"stages"`
	TournamentId uuid.UUID `json:"tournament_id"`

	// TournamentVersion Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	TournamentVersion int `json:"tournament_version"`
}

// EmailRequest описывает модель EmailRequest.
type EmailRequest struct {
	Email string `json:"email"`
}

// File Сервер проверяет сигнатуру и размер; avatar/team_logo допускают только PNG/JPEG. Доступ к протоколу ограничен участниками, организатором и назначенным судьёй.
type File struct {
	ContentType FileContentType `json:"content_type"`
	CreatedAt   time.Time       `json:"created_at"`
	Filename    string          `json:"filename"`
	Id          uuid.UUID       `json:"id"`
	Purpose     FilePurpose     `json:"purpose"`
	SizeBytes   int             `json:"size_bytes"`
}

// FileContentType описывает модель File.ContentType.
type FileContentType string

// FilePurpose описывает модель File.Purpose.
type FilePurpose string

// GameScore описывает модель GameScore.
type GameScore struct {
	AwayScore int `json:"away_score"`
	HomeScore int `json:"home_score"`
	Number    int `json:"number"`
}

// Group описывает модель Group.
type Group struct {
	ApplicationIds []uuid.UUID `json:"application_ids"`
	Id             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	StageId        uuid.UUID   `json:"stage_id"`
}

// GroupSettings wildcard_place обязателен при wildcard_count > 0. Число участников и выходящих проверяется при генерации.
type GroupSettings struct {
	GroupCount         int  `json:"group_count"`
	QualifiersPerGroup int  `json:"qualifiers_per_group"`
	WildcardCount      int  `json:"wildcard_count"`
	WildcardPlace      *int `json:"wildcard_place,omitempty"`
}

// Invitation описывает модель Invitation.
type Invitation struct {
	ExpiresAt time.Time        `json:"expires_at"`
	Id        uuid.UUID        `json:"id"`
	InviteeId *uuid.UUID       `json:"invitee_id,omitempty"`
	Status    InvitationStatus `json:"status"`
	TeamId    uuid.UUID        `json:"team_id"`
}

// InvitationStatus описывает модель Invitation.Status.
type InvitationStatus string

// InvitationDecisionRequest описывает модель InvitationDecisionRequest.
type InvitationDecisionRequest struct {
	Decision InvitationDecisionRequestDecision `json:"decision"`

	// Token Обязателен для приглашения по ссылке. Адресное приглашение принимает только его получатель.
	Token *string `json:"token,omitempty"`
}

// InvitationDecisionRequestDecision описывает модель InvitationDecisionRequest.Decision.
type InvitationDecisionRequestDecision string

// InvitationLink описывает модель InvitationLink.
type InvitationLink struct {
	Invitation Invitation `json:"invitation"`
	Token      string     `json:"token"`
	Url        string     `json:"url"`
}

// InvitationPage описывает модель InvitationPage.
type InvitationPage struct {
	Items  []Invitation `json:"items"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

// LoginRequest описывает модель LoginRequest.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Match Отсутствующие участники означают ожидаемый результат предыдущего матча/bye. Bye не меняет Эло. Время UTC; перенос не создаёт пересечения для участников/судьи. Связь loser_next используется для матча за 3-е место.
type Match struct {
	AwayApplicationId  *uuid.UUID               `json:"away_application_id,omitempty"`
	AwayScore          *int                     `json:"away_score,omitempty"`
	GroupId            *uuid.UUID               `json:"group_id,omitempty"`
	HomeApplicationId  *uuid.UUID               `json:"home_application_id,omitempty"`
	HomeScore          *int                     `json:"home_score,omitempty"`
	Id                 uuid.UUID                `json:"id"`
	LoserNextMatchId   *uuid.UUID               `json:"loser_next_match_id,omitempty"`
	LoserNextMatchSlot *MatchLoserNextMatchSlot `json:"loser_next_match_slot,omitempty"`
	NextMatchId        *uuid.UUID               `json:"next_match_id,omitempty"`
	NextMatchSlot      *MatchNextMatchSlot      `json:"next_match_slot,omitempty"`
	OnlineUrl          *string                  `json:"online_url,omitempty"`
	Position           int                      `json:"position"`
	RefereeId          *uuid.UUID               `json:"referee_id,omitempty"`
	Round              int                      `json:"round"`
	StageId            uuid.UUID                `json:"stage_id"`
	StartsAt           *time.Time               `json:"starts_at,omitempty"`
	Status             MatchStatus              `json:"status"`
	TournamentId       uuid.UUID                `json:"tournament_id"`
	Venue              *string                  `json:"venue,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version             int        `json:"version"`
	WinnerApplicationId *uuid.UUID `json:"winner_application_id,omitempty"`
}

// MatchLoserNextMatchSlot описывает модель Match.LoserNextMatchSlot.
type MatchLoserNextMatchSlot string

// MatchNextMatchSlot описывает модель Match.NextMatchSlot.
type MatchNextMatchSlot string

// MatchPage описывает модель MatchPage.
type MatchPage struct {
	Items  []Match `json:"items"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Total  int     `json:"total"`
}

// MatchStatus описывает модель MatchStatus.
type MatchStatus string

// ModerateTournamentRequest описывает модель ModerateTournamentRequest.
type ModerateTournamentRequest struct {
	Hidden bool   `json:"hidden"`
	Reason string `json:"reason"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// Notification описывает модель Notification.
type Notification struct {
	CreatedAt  time.Time  `json:"created_at"`
	Id         uuid.UUID  `json:"id"`
	Kind       string     `json:"kind"`
	Message    string     `json:"message"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	ResourceId *uuid.UUID `json:"resource_id,omitempty"`
}

// NotificationPage описывает модель NotificationPage.
type NotificationPage struct {
	Items  []Notification `json:"items"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
}

// Profile описывает модель Profile.
type Profile struct {
	AvatarUrl          *string   `json:"avatar_url,omitempty"`
	Blocked            bool      `json:"blocked"`
	Email              string    `json:"email"`
	EmailNotifications bool      `json:"email_notifications"`
	EmailVerified      bool      `json:"email_verified"`
	Id                 uuid.UUID `json:"id"`
	Name               *string   `json:"name,omitempty"`
	Nickname           string    `json:"nickname"`
	Roles              []Role    `json:"roles"`
	Timezone           string    `json:"timezone"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// ProfilePage описывает модель ProfilePage.
type ProfilePage struct {
	Items  []Profile `json:"items"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Total  int       `json:"total"`
}

// PublicUser описывает модель PublicUser.
type PublicUser struct {
	AvatarUrl *string   `json:"avatar_url,omitempty"`
	Id        uuid.UUID `json:"id"`
	Name      *string   `json:"name,omitempty"`
	Nickname  string    `json:"nickname"`
}

// PublicUserPage описывает модель PublicUserPage.
type PublicUserPage struct {
	Items  []PublicUser `json:"items"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

// Rating описывает модель Rating.
type Rating struct {
	DisciplineId uuid.UUID `json:"discipline_id"`
	Played       int       `json:"played"`
	Position     int       `json:"position"`
	Team         Team      `json:"team"`
	Value        float64   `json:"value"`
}

// RatingChange Начальный рейтинг 1500, K=32; дельты соперников противоположны, храним без округления. Bye исключён.
type RatingChange struct {
	After        float64   `json:"after"`
	Before       float64   `json:"before"`
	CreatedAt    time.Time `json:"created_at"`
	Delta        float64   `json:"delta"`
	DisciplineId uuid.UUID `json:"discipline_id"`
	MatchId      uuid.UUID `json:"match_id"`
	TeamId       uuid.UUID `json:"team_id"`
}

// RatingChangePage описывает модель RatingChangePage.
type RatingChangePage struct {
	Items  []RatingChange `json:"items"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
}

// RatingPage описывает модель RatingPage.
type RatingPage struct {
	Items  []Rating `json:"items"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	Total  int      `json:"total"`
}

// ReasonRequest описывает модель ReasonRequest.
type ReasonRequest struct {
	Reason string `json:"reason"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// RegisterRequest описывает модель RegisterRequest.
type RegisterRequest struct {
	Email               string                             `json:"email"`
	Name                *string                            `json:"name,omitempty"`
	Nickname            string                             `json:"nickname"`
	Password            string                             `json:"password"`
	PersonalDataConsent RegisterRequestPersonalDataConsent `json:"personal_data_consent"`
}

// RegisterRequestPersonalDataConsent описывает модель RegisterRequest.PersonalDataConsent.
type RegisterRequestPersonalDataConsent bool

// ReportResultRequest Начало матча должно наступить. Для Bo3/Bo5 итоговый счёт — выигранные партии и согласован с games; для Bo1 — счёт единственной партии. В плей-офф ничья запрещена. Вложение PDF/PNG/JPG <= 5 МиБ.
type ReportResultRequest struct {
	AwayScore       int          `json:"away_score"`
	EvidenceFileIds *[]uuid.UUID `json:"evidence_file_ids,omitempty"`
	Games           []GameScore  `json:"games"`
	HomeScore       int          `json:"home_score"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// ResetPasswordRequest описывает модель ResetPasswordRequest.
type ResetPasswordRequest struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

// ResolveDisputeRequest Назначенный судья устанавливает итог. Коррекция завершённого матча требует пересчёта последующих рейтингов; после старта зависимого матча изменение победителя запрещено (409).
type ResolveDisputeRequest struct {
	AwayScore       int          `json:"away_score"`
	EvidenceFileIds *[]uuid.UUID `json:"evidence_file_ids,omitempty"`
	Games           []GameScore  `json:"games"`
	HomeScore       int          `json:"home_score"`
	Reason          string       `json:"reason"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// ResultReport описывает модель ResultReport.
type ResultReport struct {
	AuthorId             uuid.UUID          `json:"author_id"`
	AwayScore            int                `json:"away_score"`
	ConfirmationDeadline *time.Time         `json:"confirmation_deadline,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	EvidenceFileIds      []uuid.UUID        `json:"evidence_file_ids"`
	Games                []GameScore        `json:"games"`
	HomeScore            int                `json:"home_score"`
	Id                   uuid.UUID          `json:"id"`
	MatchId              uuid.UUID          `json:"match_id"`
	Status               ResultReportStatus `json:"status"`
}

// ResultReportStatus описывает модель ResultReport.Status.
type ResultReportStatus string

// ReviewApplicationRequest При reject причина обязательна. Принятие проверяет число свободных мест в транзакции.
type ReviewApplicationRequest struct {
	Decision ReviewApplicationRequestDecision `json:"decision"`
	Reason   *string                          `json:"reason,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// ReviewApplicationRequestDecision описывает модель ReviewApplicationRequest.Decision.
type ReviewApplicationRequestDecision string

// Role описывает модель Role.
type Role string

// Rules group_settings обязателен только для groups_playoffs; third_place_match применим к плей-офф. Бо-серия: победитель выигрывает (best_of+1)/2 партий; правила ничьей проверяются по дисциплине/этапу.
type Rules struct {
	BestOf                   RulesBestOf `json:"best_of"`
	ConfirmationTimeoutHours int         `json:"confirmation_timeout_hours"`
	DrawPoints               int         `json:"draw_points"`

	// GroupSettings wildcard_place обязателен при wildcard_count > 0. Число участников и выходящих проверяется при генерации.
	GroupSettings   *GroupSettings      `json:"group_settings,omitempty"`
	LossPoints      int                 `json:"loss_points"`
	RoundRobinLegs  RulesRoundRobinLegs `json:"round_robin_legs"`
	ThirdPlaceMatch bool                `json:"third_place_match"`
	TieBreakers     []TieBreaker        `json:"tie_breakers"`
	WinPoints       int                 `json:"win_points"`
}

// RulesBestOf описывает модель Rules.BestOf.
type RulesBestOf int

// RulesRoundRobinLegs описывает модель Rules.RoundRobinLegs.
type RulesRoundRobinLegs int

// ScheduleMatchRequest Нужна площадка или онлайн-ссылка; нельзя перенести завершённый матч. Длительность слота 1 час; слот целиком находится внутри дат турнира. Зависимый матч начинается после предыдущего слота. При появлении участника конфликтующее время следующего матча сбрасывается с уведомлением.
type ScheduleMatchRequest struct {
	OnlineUrl *string   `json:"online_url,omitempty"`
	StartsAt  time.Time `json:"starts_at"`
	Venue     *string   `json:"venue,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// ScheduleRoundRequest Матчи тура назначаются по position с указанным интервалом; конфликты проверяются до записи всей пачки.
type ScheduleRoundRequest struct {
	IntervalMinutes int       `json:"interval_minutes"`
	Round           int       `json:"round"`
	StageId         uuid.UUID `json:"stage_id"`
	StartsAt        time.Time `json:"starts_at"`
	Venue           string    `json:"venue"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// Seed описывает модель Seed.
type Seed struct {
	ApplicationId uuid.UUID `json:"application_id"`
	Position      int       `json:"position"`
}

// SeedingRequest manual требует application_ids — полную перестановку принятых заявок, первая имеет seed=1. random сохраняет выбранную перестановку в предпросмотре.
type SeedingRequest struct {
	ApplicationIds *[]uuid.UUID         `json:"application_ids,omitempty"`
	Method         SeedingRequestMethod `json:"method"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// SeedingRequestMethod описывает модель SeedingRequest.Method.
type SeedingRequestMethod string

// Stage описывает модель Stage.
type Stage struct {
	Id           uuid.UUID   `json:"id"`
	Kind         StageKind   `json:"kind"`
	Position     int         `json:"position"`
	Status       StageStatus `json:"status"`
	TournamentId uuid.UUID   `json:"tournament_id"`
}

// StageStatus описывает модель Stage.Status.
type StageStatus string

// StageKind описывает модель StageKind.
type StageKind string

// Standing описывает модель Standing.
type Standing struct {
	ApplicationId   uuid.UUID `json:"application_id"`
	Draws           int       `json:"draws"`
	Losses          int       `json:"losses"`
	Played          int       `json:"played"`
	Points          int       `json:"points"`
	Position        int       `json:"position"`
	ScoreAgainst    int       `json:"score_against"`
	ScoreDifference int       `json:"score_difference"`
	ScoreFor        int       `json:"score_for"`
	Wins            int       `json:"wins"`
}

// Standings описывает модель Standings.
type Standings struct {
	GroupId *uuid.UUID `json:"group_id,omitempty"`
	Rows    []Standing `json:"rows"`
	StageId uuid.UUID  `json:"stage_id"`
}

// Team описывает модель Team.
type Team struct {
	CaptainId   uuid.UUID `json:"captain_id"`
	Description *string   `json:"description,omitempty"`
	Id          uuid.UUID `json:"id"`
	LogoUrl     *string   `json:"logo_url,omitempty"`
	Name        string    `json:"name"`
	Personal    bool      `json:"personal"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// TeamMember описывает модель TeamMember.
type TeamMember struct {
	Captain  bool       `json:"captain"`
	JoinedAt time.Time  `json:"joined_at"`
	User     PublicUser `json:"user"`
}

// TeamPage описывает модель TeamPage.
type TeamPage struct {
	Items  []Team `json:"items"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Total  int    `json:"total"`
}

// TieBreaker описывает модель TieBreaker.
type TieBreaker string

// TokenRequest описывает модель TokenRequest.
type TokenRequest struct {
	Token string `json:"token"`
}

// Tokens Access: 15 минут; refresh: 30 дней. Refresh ротируется и отзывается при logout/reset password; сервер хранит только хеш.
type Tokens struct {
	AccessToken  string          `json:"access_token"`
	ExpiresIn    TokensExpiresIn `json:"expires_in"`
	RefreshToken string          `json:"refresh_token"`
	TokenType    TokensTokenType `json:"token_type"`
}

// TokensExpiresIn описывает модель Tokens.ExpiresIn.
type TokensExpiresIn int

// TokensTokenType описывает модель Tokens.TokenType.
type TokensTokenType string

// Tournament описывает модель Tournament.
type Tournament struct {
	AcceptedCount        int              `json:"accepted_count"`
	CreatedAt            time.Time        `json:"created_at"`
	Description          *string          `json:"description,omitempty"`
	DisciplineId         uuid.UUID        `json:"discipline_id"`
	EndsAt               *time.Time       `json:"ends_at,omitempty"`
	Format               TournamentFormat `json:"format"`
	Hidden               bool             `json:"hidden"`
	Id                   uuid.UUID        `json:"id"`
	MaxParticipants      int              `json:"max_participants"`
	Name                 string           `json:"name"`
	Organizer            PublicUser       `json:"organizer"`
	RefereeId            *uuid.UUID       `json:"referee_id,omitempty"`
	RegistrationClosesAt *time.Time       `json:"registration_closes_at,omitempty"`
	RegistrationOpensAt  *time.Time       `json:"registration_opens_at,omitempty"`

	// Rules group_settings обязателен только для groups_playoffs; third_place_match применим к плей-офф. Бо-серия: победитель выигрывает (best_of+1)/2 партий; правила ничьей проверяются по дисциплине/этапу.
	Rules    Rules            `json:"rules"`
	StartsAt *time.Time       `json:"starts_at,omitempty"`
	Status   TournamentStatus `json:"status"`
	Timezone string           `json:"timezone"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// TournamentFormat описывает модель TournamentFormat.
type TournamentFormat string

// TournamentPage описывает модель TournamentPage.
type TournamentPage struct {
	Items  []Tournament `json:"items"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

// TournamentStatus описывает модель TournamentStatus.
type TournamentStatus string

// TransferCaptainRequest описывает модель TransferCaptainRequest.
type TransferCaptainRequest struct {
	UserId uuid.UUID `json:"user_id"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// UpdateDisciplineRequest описывает модель UpdateDisciplineRequest.
type UpdateDisciplineRequest struct {
	Active        *bool   `json:"active,omitempty"`
	AllowsDraw    *bool   `json:"allows_draw,omitempty"`
	MaxRosterSize *int    `json:"max_roster_size,omitempty"`
	MinRosterSize *int    `json:"min_roster_size,omitempty"`
	Name          *string `json:"name,omitempty"`
	ScoreUnit     *string `json:"score_unit,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// UpdateProfileRequest описывает модель UpdateProfileRequest.
type UpdateProfileRequest struct {
	AvatarFileId       *uuid.UUID `json:"avatar_file_id,omitempty"`
	EmailNotifications *bool      `json:"email_notifications,omitempty"`
	Name               *string    `json:"name,omitempty"`
	Nickname           *string    `json:"nickname,omitempty"`

	// Timezone Идентификатор IANA.
	Timezone *string `json:"timezone,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// UpdateRolesRequest Администратор назначает роли. Нельзя лишить систему последнего активного администратора.
type UpdateRolesRequest struct {
	Roles []Role `json:"roles"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// UpdateTeamRequest описывает модель UpdateTeamRequest.
type UpdateTeamRequest struct {
	Description *string    `json:"description,omitempty"`
	LogoFileId  *uuid.UUID `json:"logo_file_id,omitempty"`
	Name        *string    `json:"name,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// UpdateTournamentRequest Поля изменяются в черновике/регистрации; формат и состав нельзя менять после старта. max_participants не ниже числа принятых заявок.
type UpdateTournamentRequest struct {
	Description          *string           `json:"description,omitempty"`
	DisciplineId         *uuid.UUID        `json:"discipline_id,omitempty"`
	EndsAt               *time.Time        `json:"ends_at,omitempty"`
	Format               *TournamentFormat `json:"format,omitempty"`
	MaxParticipants      *int              `json:"max_participants,omitempty"`
	Name                 *string           `json:"name,omitempty"`
	RegistrationClosesAt *time.Time        `json:"registration_closes_at,omitempty"`
	RegistrationOpensAt  *time.Time        `json:"registration_opens_at,omitempty"`

	// Rules group_settings обязателен только для groups_playoffs; third_place_match применим к плей-офф. Бо-серия: победитель выигрывает (best_of+1)/2 партий; правила ничьей проверяются по дисциплине/этапу.
	Rules    *Rules     `json:"rules,omitempty"`
	StartsAt *time.Time `json:"starts_at,omitempty"`
	Timezone *string    `json:"timezone,omitempty"`

	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// VersionRequest описывает модель VersionRequest.
type VersionRequest struct {
	// Version Версия записи для оптимистической блокировки. Устаревшая версия даёт 409.
	Version int `json:"version"`
}

// Limit описывает модель Limit.
type Limit = int

// Offset описывает модель Offset.
type Offset = int

// ListAuditParams содержит параметры операции ListAudit.
type ListAuditParams struct {
	Limit    *Limit     `form:"limit,omitempty" json:"limit,omitempty"`
	Offset   *Offset    `form:"offset,omitempty" json:"offset,omitempty"`
	EntityId *uuid.UUID `form:"entity_id,omitempty" json:"entity_id,omitempty"`
	ActorId  *uuid.UUID `form:"actor_id,omitempty" json:"actor_id,omitempty"`
}

// ListAdminTournamentsParams содержит параметры операции ListAdminTournaments.
type ListAdminTournamentsParams struct {
	Limit        *Limit            `form:"limit,omitempty" json:"limit,omitempty"`
	Offset       *Offset           `form:"offset,omitempty" json:"offset,omitempty"`
	DisciplineId *uuid.UUID        `form:"discipline_id,omitempty" json:"discipline_id,omitempty"`
	Status       *TournamentStatus `form:"status,omitempty" json:"status,omitempty"`
	Format       *TournamentFormat `form:"format,omitempty" json:"format,omitempty"`
	StartsFrom   *time.Time        `form:"starts_from,omitempty" json:"starts_from,omitempty"`
	StartsTo     *time.Time        `form:"starts_to,omitempty" json:"starts_to,omitempty"`
	Search       *string           `form:"search,omitempty" json:"search,omitempty"`
}

// ListAdminUsersParams содержит параметры операции ListAdminUsers.
type ListAdminUsersParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
	Search *string `form:"search,omitempty" json:"search,omitempty"`
}

// ListDisciplinesParams содержит параметры операции ListDisciplines.
type ListDisciplinesParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
}

// ListAssignedDisputesParams содержит параметры операции ListAssignedDisputes.
type ListAssignedDisputesParams struct {
	Limit  *Limit                            `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset                           `form:"offset,omitempty" json:"offset,omitempty"`
	Status *ListAssignedDisputesParamsStatus `form:"status,omitempty" json:"status,omitempty"`
}

// ListAssignedDisputesParamsStatus содержит параметры операции ListAssignedDisputes.
type ListAssignedDisputesParamsStatus string

// ListRatingsParams содержит параметры операции ListRatings.
type ListRatingsParams struct {
	Limit        *Limit    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset       *Offset   `form:"offset,omitempty" json:"offset,omitempty"`
	DisciplineId uuid.UUID `form:"discipline_id" json:"discipline_id"`
}

// ListTeamsParams содержит параметры операции ListTeams.
type ListTeamsParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
	Search *string `form:"search,omitempty" json:"search,omitempty"`
}

// ListRatingHistoryParams содержит параметры операции ListRatingHistory.
type ListRatingHistoryParams struct {
	Limit        *Limit    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset       *Offset   `form:"offset,omitempty" json:"offset,omitempty"`
	DisciplineId uuid.UUID `form:"discipline_id" json:"discipline_id"`
}

// ListTournamentsParams содержит параметры операции ListTournaments.
type ListTournamentsParams struct {
	Limit        *Limit            `form:"limit,omitempty" json:"limit,omitempty"`
	Offset       *Offset           `form:"offset,omitempty" json:"offset,omitempty"`
	DisciplineId *uuid.UUID        `form:"discipline_id,omitempty" json:"discipline_id,omitempty"`
	Status       *TournamentStatus `form:"status,omitempty" json:"status,omitempty"`
	Format       *TournamentFormat `form:"format,omitempty" json:"format,omitempty"`
	StartsFrom   *time.Time        `form:"starts_from,omitempty" json:"starts_from,omitempty"`
	StartsTo     *time.Time        `form:"starts_to,omitempty" json:"starts_to,omitempty"`
	Search       *string           `form:"search,omitempty" json:"search,omitempty"`
}

// ListApplicationsParams содержит параметры операции ListApplications.
type ListApplicationsParams struct {
	Limit  *Limit             `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset            `form:"offset,omitempty" json:"offset,omitempty"`
	Status *ApplicationStatus `form:"status,omitempty" json:"status,omitempty"`
}

// ListMatchesParams содержит параметры операции ListMatches.
type ListMatchesParams struct {
	Limit   *Limit       `form:"limit,omitempty" json:"limit,omitempty"`
	Offset  *Offset      `form:"offset,omitempty" json:"offset,omitempty"`
	StageId *uuid.UUID   `form:"stage_id,omitempty" json:"stage_id,omitempty"`
	GroupId *uuid.UUID   `form:"group_id,omitempty" json:"group_id,omitempty"`
	Round   *int         `form:"round,omitempty" json:"round,omitempty"`
	Status  *MatchStatus `form:"status,omitempty" json:"status,omitempty"`
}

// ListParticipantsParams содержит параметры операции ListParticipants.
type ListParticipantsParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
}

// SearchUsersParams содержит параметры операции SearchUsers.
type SearchUsersParams struct {
	Limit    *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset   *Offset `form:"offset,omitempty" json:"offset,omitempty"`
	Nickname string  `form:"nickname" json:"nickname"`
}

// ListMyApplicationsParams содержит параметры операции ListMyApplications.
type ListMyApplicationsParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
}

// ListMyInvitationsParams содержит параметры операции ListMyInvitations.
type ListMyInvitationsParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
}

// ListNotificationsParams содержит параметры операции ListNotifications.
type ListNotificationsParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
	Unread *bool   `form:"unread,omitempty" json:"unread,omitempty"`
}

// ListMyTeamsParams содержит параметры операции ListMyTeams.
type ListMyTeamsParams struct {
	Limit  *Limit  `form:"limit,omitempty" json:"limit,omitempty"`
	Offset *Offset `form:"offset,omitempty" json:"offset,omitempty"`
}

// ListMyTournamentsParams содержит параметры операции ListMyTournaments.
type ListMyTournamentsParams struct {
	Limit        *Limit            `form:"limit,omitempty" json:"limit,omitempty"`
	Offset       *Offset           `form:"offset,omitempty" json:"offset,omitempty"`
	DisciplineId *uuid.UUID        `form:"discipline_id,omitempty" json:"discipline_id,omitempty"`
	Status       *TournamentStatus `form:"status,omitempty" json:"status,omitempty"`
	Format       *TournamentFormat `form:"format,omitempty" json:"format,omitempty"`
	StartsFrom   *time.Time        `form:"starts_from,omitempty" json:"starts_from,omitempty"`
	StartsTo     *time.Time        `form:"starts_to,omitempty" json:"starts_to,omitempty"`
	Search       *string           `form:"search,omitempty" json:"search,omitempty"`
}
