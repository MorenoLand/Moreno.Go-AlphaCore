package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

const PlayerFieldCount = 634

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

func SetUint64(values []uint32, index int, value uint64) {
	values[index] = uint32(value)
	values[index+1] = uint32(value >> 32)
}

func SetFloat(values []uint32, index int, value float32) { values[index] = math.Float32bits(value) }
