package realm

import (
	"database/sql"
	"fmt"
)

type Group struct {
	ID, LeaderGUID, LootMaster int64
	LootMethod                 int64
}

func (s *Store) CreateGroup(leaderGUID int64) (Group, error) {
	result, err := s.db.Exec(`INSERT INTO "group" (leader_guid, loot_method, loot_master) VALUES (?, 0, 0)`, leaderGUID)
	if err != nil {
		return Group{}, fmt.Errorf("create group: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Group{}, fmt.Errorf("read group id: %w", err)
	}
	return Group{ID: id, LeaderGUID: leaderGUID}, nil
}

func (s *Store) AddGroupMember(groupID, guid int64) error {
	_, err := s.db.Exec(`INSERT INTO group_member (group_id, guid) VALUES (?, ?)`, groupID, guid)
	return err
}

func (s *Store) DeleteGroupMember(groupID, guid int64) error {
	_, err := s.db.Exec(`DELETE FROM group_member WHERE group_id = ? AND guid = ?`, groupID, guid)
	return err
}

func (s *Store) GroupByPlayer(guid int64) (Group, bool, error) {
	var group Group
	err := s.db.QueryRow(`SELECT g.group_id, g.leader_guid, g.loot_method, g.loot_master FROM "group" g JOIN group_member m ON m.group_id = g.group_id WHERE m.guid = ? LIMIT 1`, guid).Scan(&group.ID, &group.LeaderGUID, &group.LootMethod, &group.LootMaster)
	if err == sql.ErrNoRows {
		return Group{}, false, nil
	}
	if err != nil {
		return Group{}, false, fmt.Errorf("query player group: %w", err)
	}
	return group, true, nil
}

func (s *Store) GroupMembers(groupID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT guid FROM group_member WHERE group_id = ? ORDER BY rowid`, groupID)
	if err != nil {
		return nil, fmt.Errorf("query group members: %w", err)
	}
	defer rows.Close()
	var members []int64
	for rows.Next() {
		var guid int64
		if err := rows.Scan(&guid); err != nil {
			return nil, fmt.Errorf("scan group member: %w", err)
		}
		members = append(members, guid)
	}
	return members, rows.Err()
}

func (s *Store) UpdateGroupLeader(groupID, leaderGUID int64) error {
	_, err := s.db.Exec(`UPDATE "group" SET leader_guid = ? WHERE group_id = ?`, leaderGUID, groupID)
	return err
}

func (s *Store) DeleteGroup(groupID int64) error {
	_, err := s.db.Exec(`DELETE FROM "group" WHERE group_id = ?`, groupID)
	return err
}
