package realm

import (
	"database/sql"
	"fmt"
)

type QuestState struct {
	GUID, Quest, State, Timer int64
	Rewarded, Explored        bool
}

func (s *Store) QuestState(guid, quest int64) (QuestState, bool, error) {
	var state QuestState
	var rewarded, explored int
	err := s.db.QueryRow(`SELECT guid, quest, state, rewarded, explored, timer FROM character_quest_state WHERE guid = ? AND quest = ?`, guid, quest).Scan(&state.GUID, &state.Quest, &state.State, &rewarded, &explored, &state.Timer)
	if err == sql.ErrNoRows {
		return QuestState{}, false, nil
	}
	if err != nil {
		return QuestState{}, false, fmt.Errorf("query quest state: %w", err)
	}
	state.Rewarded, state.Explored = rewarded != 0, explored != 0
	return state, true, nil
}

func (s *Store) QuestStates(guid int64) ([]QuestState, error) {
	rows, err := s.db.Query(`SELECT guid, quest, state, rewarded, explored, timer FROM character_quest_state WHERE guid = ? ORDER BY quest`, guid)
	if err != nil {
		return nil, fmt.Errorf("query quest states: %w", err)
	}
	defer rows.Close()
	states := make([]QuestState, 0)
	for rows.Next() {
		var state QuestState
		var rewarded, explored int
		if err := rows.Scan(&state.GUID, &state.Quest, &state.State, &rewarded, &explored, &state.Timer); err != nil {
			return nil, fmt.Errorf("scan quest state: %w", err)
		}
		state.Rewarded, state.Explored = rewarded != 0, explored != 0
		states = append(states, state)
	}
	return states, rows.Err()
}

func (s *Store) SaveQuestState(state QuestState) error {
	_, err := s.db.Exec(`INSERT INTO character_quest_state (guid, quest, state, rewarded, explored, timer) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(guid, quest) DO UPDATE SET state = excluded.state, rewarded = excluded.rewarded, explored = excluded.explored, timer = excluded.timer`, state.GUID, state.Quest, state.State, state.Rewarded, state.Explored, state.Timer)
	return err
}

func (s *Store) DeleteQuestState(guid, quest int64) error {
	_, err := s.db.Exec(`DELETE FROM character_quest_state WHERE guid = ? AND quest = ?`, guid, quest)
	return err
}
