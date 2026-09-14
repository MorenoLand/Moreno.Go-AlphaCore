package world

import "Moreno.AlphaCore/database/world"

func (s *WorldServer) pickpocketSpell(cast *spellCast, target *creatureState) {
	if cast == nil || target == nil || s.WorldData == nil || target.Health <= 0 || target.Template.PickpocketLootID <= 0 {
		return
	}
	responses, err := s.sendLoot(cast.caster, target.GUID, lootPickpocketSource, target.Template.PickpocketLootID, world.GameObjectTemplate{}, 0)
	if err != nil {
		return
	}
	for _, response := range responses {
		s.sendPlayer(cast.caster.GUID, response)
	}
}
