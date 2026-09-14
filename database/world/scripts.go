package world

import "fmt"

type EventScript struct {
	ID, Delay, Priority, Command int64
	DataLong                     [4]int64
	TargetParam                  [2]int64
	TargetType, DataFlags        int64
	DataInt                      [4]int64
	X, Y, Z, O                   float32
	ConditionID                  int64
	Comments                     string
}

func (s *Store) EventScripts(id int64) ([]EventScript, error) {
	rows, err := s.db.Query(`SELECT id, delay, priority, command, datalong, datalong2, datalong3, datalong4, target_param1, target_param2, target_type, data_flags, dataint, dataint2, dataint3, dataint4, x, y, z, o, condition_id, comments FROM event_scripts WHERE id = ? ORDER BY delay, priority`, id)
	if err != nil {
		return nil, fmt.Errorf("query event scripts: %w", err)
	}
	defer rows.Close()
	scripts := make([]EventScript, 0)
	for rows.Next() {
		var script EventScript
		values := []any{&script.ID, &script.Delay, &script.Priority, &script.Command}
		for index := range script.DataLong {
			values = append(values, &script.DataLong[index])
		}
		for index := range script.TargetParam {
			values = append(values, &script.TargetParam[index])
		}
		values = append(values, &script.TargetType, &script.DataFlags)
		for index := range script.DataInt {
			values = append(values, &script.DataInt[index])
		}
		values = append(values, &script.X, &script.Y, &script.Z, &script.O, &script.ConditionID, &script.Comments)
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("scan event script: %w", err)
		}
		scripts = append(scripts, script)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read event scripts: %w", err)
	}
	return scripts, nil
}
