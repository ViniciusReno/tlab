package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/portfolio"
)

func (s *Store) Positions(ctx context.Context) ([]portfolio.Position, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,bond_id,quantity,purchase_date,purchase_pu,purchase_yield,invested_brl,note FROM portfolio_positions ORDER BY created_at,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []portfolio.Position
	for rows.Next() {
		var p portfolio.Position
		var date sql.NullString
		if err := rows.Scan(&p.ID, &p.BondID, &p.Quantity, &date, &p.PurchasePU, &p.PurchaseYield, &p.InvestedBRL, &p.Note); err != nil {
			return nil, err
		}
		if date.Valid {
			d, err := bond.ParseDate(date.String)
			if err != nil {
				return nil, err
			}
			p.PurchaseDate = &d
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SavePosition receives normalized inputs from the application boundary.
func (s *Store) SavePosition(ctx context.Context, p portfolio.Position, create bool) error {
	var date any
	if p.PurchaseDate != nil {
		date = p.PurchaseDate.Format(time.DateOnly)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if create {
		_, err := s.db.ExecContext(ctx, `INSERT INTO portfolio_positions(id,bond_id,quantity,purchase_date,purchase_pu,purchase_yield,invested_brl,note,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, p.ID, p.BondID, p.Quantity, date, p.PurchasePU, p.PurchaseYield, p.InvestedBRL, p.Note, now, now)
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE portfolio_positions SET bond_id=?,quantity=?,purchase_date=?,purchase_pu=?,purchase_yield=?,invested_brl=?,note=?,updated_at=? WHERE id=?`, p.BondID, p.Quantity, date, p.PurchasePU, p.PurchaseYield, p.InvestedBRL, p.Note, now, p.ID)
	return changed(result, err)
}
func (s *Store) DeletePosition(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM portfolio_positions WHERE id=?`, id)
	return changed(result, err)
}
func changed(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n != 1 {
		return bond.InvalidInput
	}
	return err
}
