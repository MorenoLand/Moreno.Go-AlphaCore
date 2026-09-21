package world

import (
	"context"
	"testing"

	"Moreno.AlphaCore/database"
)

func TestQuestScriptData(t *testing.T) {
	databases, err := database.OpenMemory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer databases.Close()
	if _, err := databases.DB(database.World).Exec(`INSERT INTO quest_template (entry, StartScript, CompleteScript) VALUES (700, 701, 702); INSERT INTO quest_start_scripts (id, delay, command, datalong) VALUES (701, 3, 10, 200); INSERT INTO quest_end_scripts (id, delay, command, datalong) VALUES (702, 4, 15, 900)`); err != nil {
		t.Fatal(err)
	}
	quest, found, err := NewStore(databases).QuestTemplate(700)
	if err != nil || !found || quest.StartScript != 701 || quest.CompleteScript != 702 {
		t.Fatalf("quest=%#v found=%v err=%v", quest, found, err)
	}
	start, err := NewStore(databases).QuestStartScripts(701)
	if err != nil || len(start) != 1 || start[0].Delay != 3 || start[0].Command != 10 || start[0].DataLong[0] != 200 {
		t.Fatalf("start=%#v err=%v", start, err)
	}
	end, err := NewStore(databases).QuestEndScripts(702)
	if err != nil || len(end) != 1 || end[0].Delay != 4 || end[0].Command != 15 || end[0].DataLong[0] != 900 {
		t.Fatalf("end=%#v err=%v", end, err)
	}
}
