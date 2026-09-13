package dbc

import (
	"database/sql"
	"fmt"
)

func (s *Store) BankSlotCost(slot int64) (int64, bool, error) {
	var cost int64
	err := s.db.QueryRow(`SELECT Cost FROM BankBagSlotPrices WHERE ID = ?`, slot).Scan(&cost)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("query bank slot cost: %w", err)
	}
	return cost, true, nil
}
