package world

import "Moreno.AlphaCore/database/realm"

func (s *WorldServer) taxiEnableAll(active *realm.Character, enable bool, gmLevel int) error {
	if active == nil || gmLevel <= 0 || s.DBC == nil || s.Characters == nil {
		return nil
	}
	mask := uint64(0)
	if enable {
		nodes, err := s.DBC.TaxiNodesAll()
		if err != nil {
			return err
		}
		team := int64(0)
		if race, found, err := s.DBC.Race(active.Race); err != nil {
			return err
		} else if found {
			switch race.BaseLanguage {
			case 1:
				team = 67
			case 7:
				team = 469
			}
		}
		for _, node := range nodes {
			if node.Team == team {
				mask |= taxiMaskBit(node.ID)
			}
		}
	}
	active.Taximask = taxiMaskString(mask)
	if err := s.Characters.UpdateTaximask(active.GUID, active.AccountID, active.RealmID, active.Taximask); err != nil {
		return err
	}
	s.updatePlayer(*active)
	return nil
}
