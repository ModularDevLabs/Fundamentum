package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ModularDevLabs/Fundamentum/internal/models"
)

type Web3ScansRepo struct {
	db *sql.DB
}

func (r *Web3ScansRepo) GetOrCreateFirstScan(ctx context.Context, row models.Web3FirstScanRow) (models.Web3FirstScanRow, bool, error) {
	if row.FirstScannedAt.IsZero() {
		row.FirstScannedAt = time.Now().UTC()
	}
	if row.FirstScannerName == "" {
		row.FirstScannerName = row.FirstScannerUserID
	}
	res, err := r.db.ExecContext(ctx, `INSERT INTO web3_first_scans(
		guild_id, asset_key, asset_type, display_symbol, display_name, first_scanner_user_id, first_scanner_name, first_price_usd, first_scanned_at
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guild_id, asset_key) DO NOTHING`,
		row.GuildID,
		row.AssetKey,
		row.AssetType,
		row.DisplaySymbol,
		row.DisplayName,
		row.FirstScannerUserID,
		row.FirstScannerName,
		row.FirstPriceUSD,
		row.FirstScannedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return models.Web3FirstScanRow{}, false, err
	}
	created := false
	if n, err := res.RowsAffected(); err == nil && n > 0 {
		created = true
	}
	out, ok, err := r.GetFirstScan(ctx, row.GuildID, row.AssetKey)
	if err != nil {
		return models.Web3FirstScanRow{}, false, err
	}
	if !ok {
		return models.Web3FirstScanRow{}, false, sql.ErrNoRows
	}
	return out, created, nil
}

func (r *Web3ScansRepo) GetFirstScan(ctx context.Context, guildID, assetKey string) (models.Web3FirstScanRow, bool, error) {
	var out models.Web3FirstScanRow
	var scannedRaw string
	err := r.db.QueryRowContext(ctx, `SELECT guild_id, asset_key, asset_type, display_symbol, display_name, first_scanner_user_id, first_scanner_name, first_price_usd, first_scanned_at
		FROM web3_first_scans WHERE guild_id = ? AND asset_key = ?`,
		guildID, assetKey,
	).Scan(
		&out.GuildID,
		&out.AssetKey,
		&out.AssetType,
		&out.DisplaySymbol,
		&out.DisplayName,
		&out.FirstScannerUserID,
		&out.FirstScannerName,
		&out.FirstPriceUSD,
		&scannedRaw,
	)
	if err == sql.ErrNoRows {
		return models.Web3FirstScanRow{}, false, nil
	}
	if err != nil {
		return models.Web3FirstScanRow{}, false, err
	}
	out.FirstScannedAt, _ = time.Parse(time.RFC3339, scannedRaw)
	return out, true, nil
}
