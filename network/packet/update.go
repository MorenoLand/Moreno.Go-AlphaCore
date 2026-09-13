package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

const PlayerFieldCount = 634
const UnitFieldCount = 184

const (
	UpdatePartial      byte = 0
	UpdateMovement     byte = 1
	UpdateCreateObject byte = 2
)

type Movement struct {
	X, Y, Z, O          float32
	WalkSpeed, RunSpeed float32
	SwimSpeed, TurnRate float32
	MovementFlags       uint32
}

func EncodePlayerCreate(guid uint64, values []uint32, movement Movement) ([]byte, error) {
	if len(values) != PlayerFieldCount {
		return nil, fmt.Errorf("player field count: got %d, want %d", len(values), PlayerFieldCount)
	}
	var data bytes.Buffer
	write := func(value any) { binary.Write(&data, binary.LittleEndian, value) }
	data.WriteByte(UpdateCreateObject)
	write(guid)
	data.WriteByte(4)
	write(uint64(0))
	write([4]float32{})
	write([4]float32{movement.X, movement.Y, movement.Z, movement.O})
	write(float32(0))
	write(movement.MovementFlags)
	write(uint32(0))
	write(movement.WalkSpeed)
	write(movement.RunSpeed)
	write(movement.SwimSpeed)
	write(movement.TurnRate)
	write(uint32(1))
	write(uint32(1))
	write(uint32(0))
	write(uint64(0))
	blockCount := (len(values) + 31) / 32
	data.WriteByte(byte(blockCount))
	for block := 0; block < blockCount; block++ {
		var mask uint32
		for bit := 0; bit < 32; bit++ {
			index := block*32 + bit
			if index < len(values) {
				mask |= 1 << bit
			}
		}
		write(mask)
	}
	for _, value := range values {
		write(value)
	}
	body := make([]byte, 4+data.Len())
	binary.LittleEndian.PutUint32(body, 1)
	copy(body[4:], data.Bytes())
	return Encode(SMSGUpdateObject, body)
}

func EncodeItemCreate(guid uint64, entry uint32, owner, creator uint64, stack uint32, duration int32, flags uint32, charges [5]int64, movement Movement) ([]byte, error) {
	var data bytes.Buffer
	write := func(value any) { binary.Write(&data, binary.LittleEndian, value) }
	data.WriteByte(UpdateCreateObject)
	write(guid)
	data.WriteByte(1)
	write(uint64(0))
	write([4]float32{})
	write([4]float32{movement.X, movement.Y, movement.Z, movement.O})
	write(float32(0))
	write(movement.MovementFlags)
	write(uint32(0))
	write(float32(1))
	write(float32(1))
	write(float32(1))
	write(float32(1))
	write(uint32(1))
	write(uint32(1))
	write(uint32(0))
	write(uint64(0))
	values := make([]uint32, 36)
	SetUint64(values, 0, guid)
	values[2] = 3
	values[3] = entry
	values[4] = math.Float32bits(1)
	SetUint64(values, 6, owner)
	SetUint64(values, 8, owner)
	SetUint64(values, 10, creator)
	values[12] = stack
	values[13] = uint32(duration)
	hasCharges := false
	for _, charge := range charges {
		if charge != 0 {
			hasCharges = true
		}
	}
	for index, charge := range charges {
		if hasCharges {
			values[14+index] = uint32(charge)
		} else {
			values[14+index] = ^uint32(0)
		}
	}
	values[19] = flags
	data.WriteByte(2)
	write(uint32(0xffffffff))
	write(uint32(0xf))
	for _, value := range values {
		write(value)
	}
	body := make([]byte, 4+data.Len())
	binary.LittleEndian.PutUint32(body, 1)
	copy(body[4:], data.Bytes())
	return Encode(SMSGUpdateObject, body)
}

func EncodeUnitCreate(guid uint64, fields []uint32, movement Movement) ([]byte, error) {
	if len(fields) != UnitFieldCount {
		return nil, fmt.Errorf("unit field count: got %d, want %d", len(fields), UnitFieldCount)
	}
	var data bytes.Buffer
	write := func(value any) { binary.Write(&data, binary.LittleEndian, value) }
	data.WriteByte(UpdateCreateObject)
	write(guid)
	data.WriteByte(3)
	write(uint64(0))
	write([4]float32{})
	write([4]float32{movement.X, movement.Y, movement.Z, movement.O})
	write(float32(0))
	write(movement.MovementFlags)
	write(uint32(0))
	write(movement.WalkSpeed)
	write(movement.RunSpeed)
	write(movement.SwimSpeed)
	write(movement.TurnRate)
	write(uint32(0))
	write(uint32(0))
	write(uint32(0))
	write(uint64(0))
	blockCount := (len(fields) + 31) / 32
	data.WriteByte(byte(blockCount))
	for block := 0; block < blockCount; block++ {
		var mask uint32
		for bit := 0; bit < 32; bit++ {
			index := block*32 + bit
			if index < len(fields) {
				mask |= 1 << bit
			}
		}
		write(mask)
	}
	for _, value := range fields {
		write(value)
	}
	body := make([]byte, 4+data.Len())
	binary.LittleEndian.PutUint32(body, 1)
	copy(body[4:], data.Bytes())
	return Encode(SMSGUpdateObject, body)
}

func EncodeGameObjectCreate(guid uint64, fields []uint32, movement Movement) ([]byte, error) {
	if len(fields) != 20 {
		return nil, fmt.Errorf("gameobject field count: got %d, want 20", len(fields))
	}
	var data bytes.Buffer
	write := func(value any) { binary.Write(&data, binary.LittleEndian, value) }
	data.WriteByte(UpdateCreateObject)
	write(guid)
	data.WriteByte(5)
	write(uint64(0))
	write([4]float32{})
	write([4]float32{movement.X, movement.Y, movement.Z, movement.O})
	write(float32(0))
	write(movement.MovementFlags)
	write(uint32(0))
	write(float32(1))
	write(float32(1))
	write(float32(1))
	write(float32(1))
	write(uint32(0))
	write(uint32(0))
	write(uint32(0))
	write(uint64(0))
	data.WriteByte(1)
	write(uint32(0xfffff))
	for _, value := range fields {
		write(value)
	}
	body := make([]byte, 4+data.Len())
	binary.LittleEndian.PutUint32(body, 1)
	copy(body[4:], data.Bytes())
	return Encode(SMSGUpdateObject, body)
}

func SetUint64(values []uint32, index int, value uint64) {
	values[index] = uint32(value)
	values[index+1] = uint32(value >> 32)
}

func SetFloat(values []uint32, index int, value float32) { values[index] = math.Float32bits(value) }
