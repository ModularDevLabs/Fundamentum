package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type EconomyRepo struct {
	db *sql.DB
}

type EconomyBalanceRow struct {
	UserID  string `json:"user_id"`
	Balance int    `json:"balance"`
}

type ShopItemRow struct {
	ID              int64  `json:"id"`
	GuildID         string `json:"guild_id"`
	Name            string `json:"name"`
	Cost            int    `json:"cost"`
	RoleID          string `json:"role_id"`
	DurationMinutes int    `json:"duration_minutes"`
	Enabled         bool   `json:"enabled"`
	CreatedAt       string `json:"created_at"`
}

func (r *EconomyRepo) AddBalance(ctx context.Context, guildID, userID string, delta int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO economy_balances(guild_id, user_id, balance, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET balance = economy_balances.balance + excluded.balance, updated_at = excluded.updated_at`,
		guildID, userID, delta, now,
	)
	return err
}

func (r *EconomyRepo) GetBalance(ctx context.Context, guildID, userID string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(balance, 0) FROM economy_balances WHERE guild_id = ? AND user_id = ?`, guildID, userID)
	var n int
	if err := row.Scan(&n); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return n, nil
}

func (r *EconomyRepo) Leaderboard(ctx context.Context, guildID string, limit int) ([]EconomyBalanceRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, balance FROM economy_balances WHERE guild_id = ? ORDER BY balance DESC LIMIT ?`, guildID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EconomyBalanceRow, 0, limit)
	for rows.Next() {
		var row EconomyBalanceRow
		if err := rows.Scan(&row.UserID, &row.Balance); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *EconomyRepo) ListShopItems(ctx context.Context, guildID string) ([]ShopItemRow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, guild_id, name, cost, role_id, duration_minutes, enabled, created_at
		FROM shop_items WHERE guild_id = ? ORDER BY cost ASC, id ASC`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ShopItemRow, 0, 32)
	for rows.Next() {
		var row ShopItemRow
		if err := rows.Scan(&row.ID, &row.GuildID, &row.Name, &row.Cost, &row.RoleID, &row.DurationMinutes, &row.Enabled, &row.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *EconomyRepo) AddShopItem(ctx context.Context, row ShopItemRow) (int64, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `INSERT INTO shop_items(guild_id, name, cost, role_id, duration_minutes, enabled, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)`,
		row.GuildID, row.Name, row.Cost, row.RoleID, row.DurationMinutes, row.Enabled, created,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *EconomyRepo) GetShopItem(ctx context.Context, guildID string, itemID int64) (ShopItemRow, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, guild_id, name, cost, role_id, duration_minutes, enabled, created_at
		FROM shop_items WHERE guild_id = ? AND id = ?`, guildID, itemID)
	var out ShopItemRow
	if err := row.Scan(&out.ID, &out.GuildID, &out.Name, &out.Cost, &out.RoleID, &out.DurationMinutes, &out.Enabled, &out.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return ShopItemRow{}, false, nil
		}
		return ShopItemRow{}, false, err
	}
	return out, true, nil
}

func (r *EconomyRepo) ResetBalances(ctx context.Context, guildID string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM economy_balances WHERE guild_id = ?`, guildID)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("economy reset rows affected: %w", err)
	}
	return n, nil
}
