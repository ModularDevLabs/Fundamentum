package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Keep one shared connection for SQLite to avoid lock churn across pooled conns.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, fmt.Errorf("set journal_mode WAL: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}
	if _, err := db.Exec(`PRAGMA synchronous=NORMAL;`); err != nil {
		return nil, fmt.Errorf("set synchronous: %w", err)
	}

	return db, nil
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Activity:        &ActivityRepo{db: db},
		Actions:         &ActionsRepo{db: db},
		Settings:        &SettingsRepo{db: db},
		Backfill:        &BackfillRepo{db: db},
		ReactionRoles:   &ReactionRolesRepo{db: db},
		Warnings:        &WarningsRepo{db: db},
		Scheduled:       &ScheduledMessagesRepo{db: db},
		Tickets:         &TicketsRepo{db: db},
		Appeals:         &AppealsRepo{db: db},
		CustomCommands:  &CustomCommandsRepo{db: db},
		Starboard:       &StarboardRepo{db: db},
		Leveling:        &LevelingRepo{db: db},
		Giveaways:       &GiveawaysRepo{db: db},
		Polls:           &PollsRepo{db: db},
		Suggestions:     &SuggestionsRepo{db: db},
		AFK:             &AFKRepo{db: db},
		Reminders:       &RemindersRepo{db: db},
		MemberNotes:     &MemberNotesRepo{db: db},
		Retention:       &RetentionRepo{db: db},
		Webhooks:        &WebhooksRepo{db: db},
		AuditTrail:      &AuditTrailRepo{db: db},
		Reputation:      &ReputationRepo{db: db},
		Economy:         &EconomyRepo{db: db},
		Achievements:    &AchievementsRepo{db: db},
		Calendar:        &CalendarRepo{db: db},
		RoleRentals:     &RoleRentalsRepo{db: db},
		Confessions:     &ConfessionsRepo{db: db},
		Trivia:          &TriviaRepo{db: db},
		Birthdays:       &BirthdaysRepo{db: db},
		RoleProgression: &RoleProgressionRepo{db: db},
		JoinScreening:   &JoinScreeningRepo{db: db},
		RaidPanic:       &RaidPanicRepo{db: db},
		Streaks:         &StreaksRepo{db: db},
	}
}

type Repositories struct {
	Activity        *ActivityRepo
	Actions         *ActionsRepo
	Settings        *SettingsRepo
	Backfill        *BackfillRepo
	ReactionRoles   *ReactionRolesRepo
	Warnings        *WarningsRepo
	Scheduled       *ScheduledMessagesRepo
	Tickets         *TicketsRepo
	Appeals         *AppealsRepo
	CustomCommands  *CustomCommandsRepo
	Starboard       *StarboardRepo
	Leveling        *LevelingRepo
	Giveaways       *GiveawaysRepo
	Polls           *PollsRepo
	Suggestions     *SuggestionsRepo
	AFK             *AFKRepo
	Reminders       *RemindersRepo
	MemberNotes     *MemberNotesRepo
	Retention       *RetentionRepo
	Webhooks        *WebhooksRepo
	AuditTrail      *AuditTrailRepo
	Reputation      *ReputationRepo
	Economy         *EconomyRepo
	Achievements    *AchievementsRepo
	Calendar        *CalendarRepo
	RoleRentals     *RoleRentalsRepo
	Confessions     *ConfessionsRepo
	Trivia          *TriviaRepo
	Birthdays       *BirthdaysRepo
	RoleProgression *RoleProgressionRepo
	JoinScreening   *JoinScreeningRepo
	RaidPanic       *RaidPanicRepo
	Streaks         *StreaksRepo
}
