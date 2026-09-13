package realm

import (
	"database/sql"
	"fmt"
)

type QuestState struct {
	GUID, Quest, State, Timer int64
	Rewarded, Explored        bool
	MobCounts, ItemCounts     [4]int64
}

func scanQuestState(row rowScanner) (QuestState, error) {
	var state QuestState
	var rewarded, explored int
	values := []interface{}{&state.GUID, &state.Quest, &state.State, &rewarded, &explored, &state.Timer}
	for index := range state.MobCounts {
		values = append(values, &state.MobCounts[index])
	}
	for index := range state.ItemCounts {
		values = append(values, &state.ItemCounts[index])
	}
	err := row.Scan(values...)
	state.Rewarded, state.Explored = rewarded != 0, explored != 0
	return state, err
}

func (s *Store) QuestState(guid, quest int64) (QuestState, bool, error) {
	state, err := scanQuestState(s.db.QueryRow(`SELECT guid, quest, state, rewarded, explored, timer, mobcount1, mobcount2, mobcount3, mobcount4, itemcount1, itemcount2, itemcount3, itemcount4 FROM character_quest_state WHERE guid = ? AND quest = ?`, guid, quest))
	if err == sql.ErrNoRows {
		return QuestState{}, false, nil
	}
	if err != nil {
		return QuestState{}, false, fmt.Errorf("query quest state: %w", err)
	}
	return state, true, nil
}

func (s *Store) QuestStates(guid int64) ([]QuestState, error) {
	rows, err := s.db.Query(`SELECT guid, quest, state, rewarded, explored, timer, mobcount1, mobcount2, mobcount3, mobcount4, itemcount1, itemcount2, itemcount3, itemcount4 FROM character_quest_state WHERE guid = ? ORDER BY quest`, guid)
	if err != nil {
		return nil, fmt.Errorf("query quest states: %w", err)
	}
	defer rows.Close()
	states := make([]QuestState, 0)
	for rows.Next() {
		state, err := scanQuestState(rows)
		if err != nil {
			return nil, fmt.Errorf("scan quest state: %w", err)
		}
		states = append(states, state)
	}
	return states, rows.Err()
}

func (s *Store) SaveQuestState(state QuestState) error {
	_, err := s.db.Exec(`INSERT INTO character_quest_state (guid, quest, state, rewarded, explored, timer, mobcount1, mobcount2, mobcount3, mobcount4, itemcount1, itemcount2, itemcount3, itemcount4) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(guid, quest) DO UPDATE SET state = excluded.state, rewarded = excluded.rewarded, explored = excluded.explored, timer = excluded.timer, mobcount1 = excluded.mobcount1, mobcount2 = excluded.mobcount2, mobcount3 = excluded.mobcount3, mobcount4 = excluded.mobcount4, itemcount1 = excluded.itemcount1, itemcount2 = excluded.itemcount2, itemcount3 = excluded.itemcount3, itemcount4 = excluded.itemcount4`, state.GUID, state.Quest, state.State, state.Rewarded, state.Explored, state.Timer, state.MobCounts[0], state.MobCounts[1], state.MobCounts[2], state.MobCounts[3], state.ItemCounts[0], state.ItemCounts[1], state.ItemCounts[2], state.ItemCounts[3])
	return err
}

func (s *Store) DeleteQuestState(guid, quest int64) error {
	_, err := s.db.Exec(`DELETE FROM character_quest_state WHERE guid = ? AND quest = ?`, guid, quest)
	return err
}
