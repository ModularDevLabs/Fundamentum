package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type ReactionRolesRepo struct {
	db *sql.DB
}

func (r *ReactionRolesRepo) ListByGuild(ctx context.Context, guildID string) ([]models.ReactionRoleRule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT rr.id, rr.guild_id, rr.channel_id, rr.message_id, rr.emoji, rr.role_id,
		COALESCE(c.group_key, ''), COALESCE(c.max_select, 0), COALESCE(c.min_select, 0),
		rr.remove_on_unreact, rr.created_at
		FROM reaction_role_rules rr
		LEFT JOIN reaction_role_rule_constraints c ON c.rule_id = rr.id
		WHERE guild_id = ?
		ORDER BY rr.id ASC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ReactionRoleRule, 0)
	for rows.Next() {
		var rule models.ReactionRoleRule
		var removeOnUnreactInt int
		var created string
		if err := rows.Scan(&rule.ID, &rule.GuildID, &rule.ChannelID, &rule.MessageID, &rule.Emoji, &rule.RoleID, &rule.GroupKey, &rule.MaxSelect, &rule.MinSelect, &removeOnUnreactInt, &created); err != nil {
			return nil, err
		}
		rule.RemoveOnUnreact = removeOnUnreactInt == 1
		rule.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *ReactionRolesRepo) Create(ctx context.Context, rule models.ReactionRoleRule) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	removeOnUnreact := 0
	if rule.RemoveOnUnreact {
		removeOnUnreact = 1
	}
	res, err := r.db.ExecContext(ctx, `INSERT INTO reaction_role_rules(
		guild_id, channel_id, message_id, emoji, role_id, remove_on_unreact, created_at
	) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		rule.GuildID, rule.ChannelID, rule.MessageID, rule.Emoji, rule.RoleID, removeOnUnreact, now,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	_, _ = r.db.ExecContext(ctx, `INSERT OR REPLACE INTO reaction_role_rule_constraints(rule_id, group_key, max_select, min_select)
		VALUES(?, ?, ?, ?)`, id, rule.GroupKey, rule.MaxSelect, rule.MinSelect)
	return id, nil
}

func (r *ReactionRolesRepo) Delete(ctx context.Context, guildID string, id int64) error {
	_, _ = r.db.ExecContext(ctx, `DELETE FROM reaction_role_rule_constraints WHERE rule_id = ?`, id)
	_, err := r.db.ExecContext(ctx, `DELETE FROM reaction_role_rules WHERE guild_id = ? AND id = ?`, guildID, id)
	return err
}

func (r *ReactionRolesRepo) DeleteAllByGuild(ctx context.Context, guildID string) error {
	_, _ = r.db.ExecContext(ctx, `DELETE FROM reaction_role_rule_constraints WHERE rule_id IN (SELECT id FROM reaction_role_rules WHERE guild_id = ?)`, guildID)
	_, err := r.db.ExecContext(ctx, `DELETE FROM reaction_role_rules WHERE guild_id = ?`, guildID)
	return err
}
