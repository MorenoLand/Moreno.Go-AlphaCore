package dbc

import (
	"database/sql"
	"fmt"
)

type TaxiPathNode struct {
	ID, PathID, NodeIndex, ContinentID int64
	X, Y, Z                            float32
	Flags                              int64
}

func (s *Store) TaxiPathNodes(pathID int64) ([]TaxiPathNode, error) {
	rows, err := s.db.Query(`SELECT ID, PathID, NodeIndex, ContinentID, LocX, LocY, LocZ, Flags FROM TaxiPathNode WHERE PathID = ? ORDER BY NodeIndex`, pathID)
	if err != nil {
		return nil, fmt.Errorf("query taxi path nodes: %w", err)
	}
	defer rows.Close()
	nodes := make([]TaxiPathNode, 0)
	for rows.Next() {
		var node TaxiPathNode
		if err := rows.Scan(&node.ID, &node.PathID, &node.NodeIndex, &node.ContinentID, &node.X, &node.Y, &node.Z, &node.Flags); err != nil {
			return nil, fmt.Errorf("scan taxi path node: %w", err)
		}
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read taxi path nodes: %w", err)
	}
	return nodes, nil
}

func (s *Store) TaxiNode(id int64) (TaxiNode, bool, error) {
	var node TaxiNode
	err := s.db.QueryRow(`SELECT ID, ContinentID, X, Y, Z, custom_Team FROM TaxiNodes WHERE ID = ?`, id).Scan(&node.ID, &node.ContinentID, &node.X, &node.Y, &node.Z, &node.Team)
	if err == nil {
		return node, true, nil
	}
	if err == sql.ErrNoRows {
		return TaxiNode{}, false, nil
	}
	return TaxiNode{}, false, fmt.Errorf("query taxi node: %w", err)
}
