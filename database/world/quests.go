package world

import (
	"database/sql"
	"fmt"
)

type QuestRelation struct {
	Entry, Quest int64
}

type QuestGreeting struct {
	Entry, Type, EmoteID, EmoteDelay int64
	Content                          string
}

func (s *Store) CreatureQuestRelations(entry int64, finisher bool) ([]QuestRelation, error) {
	table := "creature_quest_starter"
	if finisher {
		table = "creature_quest_finisher"
	}
	return s.questRelations(table, entry)
}

func (s *Store) GameObjectQuestRelations(entry int64, finisher bool) ([]QuestRelation, error) {
	table := "gameobject_quest_starter"
	if finisher {
		table = "gameobject_quest_finisher"
	}
	return s.questRelations(table, entry)
}

func (s *Store) questRelations(table string, entry int64) ([]QuestRelation, error) {
	rows, err := s.db.Query(`SELECT entry, quest FROM `+table+` WHERE entry = ? ORDER BY quest`, entry)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", table, err)
	}
	defer rows.Close()
	relations := make([]QuestRelation, 0)
	for rows.Next() {
		var relation QuestRelation
		if err := rows.Scan(&relation.Entry, &relation.Quest); err != nil {
			return nil, fmt.Errorf("scan %s: %w", table, err)
		}
		relations = append(relations, relation)
	}
	return relations, rows.Err()
}

func (s *Store) QuestGreeting(entry int64) (QuestGreeting, bool, error) {
	var greeting QuestGreeting
	err := s.db.QueryRow(`SELECT entry, type, COALESCE(content_default, ''), emote_id, emote_delay FROM quest_greeting WHERE entry = ? ORDER BY type LIMIT 1`, entry).Scan(&greeting.Entry, &greeting.Type, &greeting.Content, &greeting.EmoteID, &greeting.EmoteDelay)
	if err == sql.ErrNoRows {
		return QuestGreeting{}, false, nil
	}
	if err != nil {
		return QuestGreeting{}, false, fmt.Errorf("query quest greeting: %w", err)
	}
	return greeting, true, nil
}
