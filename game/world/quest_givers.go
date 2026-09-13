package world

import (
	"Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
)

func (s *WorldServer) questGiverEntry(active realm.Character, guid uint64) (int64, bool, bool, error) {
	_, creature, found, err := s.questGiverAt(active, guid)
	if err != nil {
		return 0, false, false, err
	}
	if found {
		return creature.Entry, false, true, nil
	}
	_, object, found, err := s.gameObjectAt(active, guid, maxShopDistance)
	if err != nil || !found {
		return 0, false, false, err
	}
	return object.Entry, true, true, nil
}

func (s *WorldServer) questRelations(entry int64, gameObject, finisher bool) ([]worlddb.QuestRelation, error) {
	if gameObject {
		return s.WorldData.GameObjectQuestRelations(entry, finisher)
	}
	return s.WorldData.CreatureQuestRelations(entry, finisher)
}
