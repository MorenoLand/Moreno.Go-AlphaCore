package world

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"Moreno.AlphaCore/database/dbc"
	"Moreno.AlphaCore/database/realm"
	"Moreno.AlphaCore/network/packet"
)

const taxiNodeMaskBits = 64

const (
	taxiOK             uint32  = 0
	taxiNoSuchPath     uint32  = 2
	taxiNotEnoughMoney uint32  = 3
	taxiTooFarAway     uint32  = 4
	taxiNoVendorNearby uint32  = 5
	taxiNotVisited     uint32  = 6
	taxiSameNode       uint32  = 11
	taxiFlightSpeed    float32 = 32
)

func (s *WorldServer) activateTaxi(active *realm.Character, data []byte) ([][]byte, error) {
	if active == nil || len(data) < 16 || s.DBC == nil || s.WorldData == nil {
		return nil, nil
	}
	guid, start, destination := binary.LittleEndian.Uint64(data), int64(binary.LittleEndian.Uint32(data[8:])), int64(binary.LittleEndian.Uint32(data[12:]))
	result := taxiOK
	_, flightMaster, found, err := s.creatureAt(*active, guid, maxShopDistance)
	if err != nil {
		return nil, err
	}
	if !found {
		result = taxiTooFarAway
	} else if flightMaster.NPCFlags&0x4 == 0 {
		result = taxiNoVendorNearby
	} else if start == destination {
		result = taxiSameNode
	} else if !taxiMaskHas(taxiMask(active.Taximask), start) || !taxiMaskHas(taxiMask(active.Taximask), destination) {
		result = taxiNotVisited
	}
	var path dbc.TaxiPath
	if result == taxiOK {
		path, found, err = s.DBC.TaxiPath(start, destination)
		if err != nil {
			return nil, err
		}
		if !found {
			result = taxiNoSuchPath
		}
	}
	var nodes []dbc.TaxiPathNode
	if result == taxiOK {
		nodes, err = s.DBC.TaxiPathNodes(path.ID)
		if err != nil {
			return nil, err
		}
		if len(nodes) == 0 {
			result = 1
		} else if active.Money < path.Cost {
			result = taxiNotEnoughMoney
		}
	}
	reply, err := packet.Encode(packet.SMSGActivateTaxiReply, encodeUint32(int64(result)))
	if err != nil {
		return nil, err
	}
	responses := [][]byte{reply}
	if result != taxiOK {
		return responses, nil
	}
	active.Money -= path.Cost
	if s.Characters != nil {
		if err := s.Characters.UpdateMoney(active.GUID, active.AccountID, active.RealmID, active.Money); err != nil {
			return nil, err
		}
	}
	active.TaxiPath = taxiPathString(nodes, start, destination, taxiMountDisplay(*active), int64(len(nodes)))
	if s.Characters != nil {
		if err := s.Characters.UpdateTaxiPath(active.GUID, active.AccountID, active.RealmID, active.TaxiPath); err != nil {
			return nil, err
		}
	}
	s.updatePlayer(*active)
	points := make([]packet.Point, 0, len(nodes))
	lastX, lastY, lastZ := active.PositionX, active.PositionY, active.PositionZ
	distance := float32(0)
	for _, node := range nodes {
		dx, dy, dz := node.X-lastX, node.Y-lastY, node.Z-lastZ
		distance += float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		points = append(points, packet.Point{X: node.X, Y: node.Y, Z: node.Z})
		lastX, lastY, lastZ = node.X, node.Y, node.Z
	}
	move, err := packet.EncodeMonsterMove(uint64(active.GUID), active.PositionX, active.PositionY, active.PositionZ, uint32(distance/taxiFlightSpeed*1000), 0x200, points)
	if err != nil {
		return nil, err
	}
	return append(responses, move), nil
}

func taxiMountDisplay(active realm.Character) int64 {
	if active.Race == 2 || active.Race == 5 || active.Race == 6 || active.Race == 8 {
		return 2157
	}
	return 1149
}

func taxiPathString(nodes []dbc.TaxiPathNode, start, destination, mount, remaining int64) string {
	return fmt.Sprintf("%f,%f,%f,%d,%d,%d,%d", nodes[0].X, nodes[0].Y, nodes[0].Z, start, destination, mount, remaining)
}

func (s *WorldServer) taxiAtDestination(active realm.Character, x, y, z float32) bool {
	parts := strings.Split(active.TaxiPath, ",")
	if len(parts) < 7 || s.DBC == nil {
		return false
	}
	destination, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return false
	}
	node, found, err := s.DBC.TaxiNode(destination)
	if err != nil || !found {
		return false
	}
	dx, dy, dz := node.X-x, node.Y-y, node.Z-z
	return dx*dx+dy*dy+dz*dz <= maxShopDistance*maxShopDistance
}

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
