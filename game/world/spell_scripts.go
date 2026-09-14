package world

import "Moreno.AlphaCore/database/realm"

func (s *WorldServer) scriptSpellEffect(cast *spellCast, target realm.Character) {
	if cast == nil || cast.spell.ID != 966 || s.Characters == nil {
		return
	}
	bind, found, err := s.Characters.Deathbind(target.GUID)
	if err != nil || !found {
		return
	}
	teleport := target
	if response, err := s.teleportPlayer(&teleport, bind.Map, bind.X, bind.Y, bind.Z, target.Orientation); err == nil {
		s.sendPlayer(target.GUID, response)
	}
}
