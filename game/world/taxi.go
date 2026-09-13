package world

import (
	"encoding/binary"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const taxiNodeMaskBits = 64

func (s *WorldServer) taxiQueryNodes(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 8 || s.DBC == nil || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	_, flightMaster, found, err := s.creatureAt(*active, guid, maxShopDistance)
	if err != nil || !found || flightMaster.NPCFlags&0x4 == 0 {
		return nil, err
	}
	nodes, err := s.DBC.TaxiNodesByMap(active.Map)
	if err != nil {
		return nil, err
	}
	current := nearestTaxiNode(nodes, active.PositionX, active.PositionY, active.PositionZ)
	if current == 0 {
		return nil, nil
	}
	known := taxiMask(active.Taximask)
	responses := make([][]byte, 0, 2)
	if !taxiMaskHas(known, current) {
		known |= taxiMaskBit(current)
		active.Taximask = taxiMaskString(known)
		if err := s.Characters.UpdateTaximask(active.GUID, active.AccountID, active.RealmID, active.Taximask); err != nil {
			return nil, err
		}
		s.updatePlayer(*active)
		newPath, err := packet.Encode(packet.SMSGNewTaxiPath, nil)
		if err != nil {
			return nil, err
		}
		status, err := packet.Encode(packet.SMSGTaxiNodeStatus, append(encodeGUID(int64(guid)), 1))
		if err != nil {
			return nil, err
		}
		return [][]byte{newPath, status}, nil
	}
	destinations := uint64(0)
	for _, node := range nodes {
		if taxiMaskHas(known, node.ID) && taxiPathExists(s, current, node.ID) {
			destinations |= taxiMaskBit(node.ID)
		}
	}
	body := append(encodeUint32(1), encodeGUID(int64(guid))...)
	body = append(body, encodeUint32(current)...)
	body = append(body, encodeUint64(destinations)...)
	body = append(body, encodeUint64(known)...)
	show, err := packet.Encode(packet.SMSGShowTaxiNodes, body)
	if err != nil {
		return nil, err
	}
	responses = append(responses, show)
	return responses, nil
}

func (s *WorldServer) taxiNodeStatus(active realm.Character, data []byte) ([]byte, error) {
	if len(data) < 8 || s.DBC == nil || s.WorldData == nil {
		return nil, nil
	}
	guid := binary.LittleEndian.Uint64(data)
	_, flightMaster, found, err := s.creatureAt(active, guid, maxShopDistance)
	if err != nil || !found || flightMaster.NPCFlags&0x4 == 0 {
		return nil, err
	}
	nodes, err := s.DBC.TaxiNodesByMap(active.Map)
	if err != nil {
		return nil, err
	}
	node := nearestTaxiNode(nodes, active.PositionX, active.PositionY, active.PositionZ)
	if node == 0 {
		return nil, nil
	}
	available := byte(1)
	if taxiMaskHas(taxiMask(active.Taximask), node) {
		available = 0
	}
	body := append(encodeGUID(int64(guid)), available)
	return packet.Encode(packet.SMSGTaxiNodeStatus, body)
}

func nearestTaxiNode(nodes []dbc.TaxiNode, x, y, z float32) int64 {
	var nearest int64
	distance := float32(-1)
	for _, node := range nodes {
		dx, dy, dz := node.X-x, node.Y-y, node.Z-z
		value := dx*dx + dy*dy + dz*dz
		if distance < 0 || value < distance {
			nearest, distance = node.ID, value
		}
	}
	return nearest
}

func taxiPathExists(s *WorldServer, from, to int64) bool {
	if _, found, err := s.DBC.TaxiPath(from, to); err == nil && found {
		return true
	}
	_, found, _ := s.DBC.TaxiPath(to, from)
	return found
}

func taxiMask(value string) uint64 {
	var mask uint64
	for index, bit := range value {
		if index >= taxiNodeMaskBits {
			break
		}
		if bit == '1' {
			mask |= uint64(1) << uint(index)
		}
	}
	return mask
}

func taxiMaskString(mask uint64) string {
	data := make([]byte, taxiNodeMaskBits)
	for index := range data {
		if mask&(uint64(1)<<uint(index)) != 0 {
			data[index] = '1'
		} else {
			data[index] = '0'
		}
	}
	return string(data)
}

func taxiMaskHas(mask uint64, node int64) bool {
	return node > 0 && node <= taxiNodeMaskBits && mask&taxiMaskBit(node) != 0
}

func taxiMaskBit(node int64) uint64 {
	if node <= 0 || node > taxiNodeMaskBits {
		return 0
	}
	return uint64(1) << uint(node-1)
}

func encodeUint64(value uint64) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, value)
	return data
}
