package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func EncodeDynamicObjectCreate(guid uint64, fields []uint32, movement Movement) ([]byte, error) {
	if len(fields) != 16 {
		return nil, fmt.Errorf("dynamic object field count: got %d, want 16", len(fields))
	}
	var data bytes.Buffer
	write := func(value any) { binary.Write(&data, binary.LittleEndian, value) }
	data.WriteByte(UpdateCreateObject)
	write(guid)
	data.WriteByte(6)
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
	write(uint32(0xffff))
	for _, value := range fields {
		write(value)
	}
	body := make([]byte, 4+data.Len())
	binary.LittleEndian.PutUint32(body, 1)
	copy(body[4:], data.Bytes())
	return Encode(SMSGUpdateObject, body)
}

func DynamicObjectFields(guid, caster uint64, spell, dynamicType int64, radius, x, y, z, orientation float32) []uint32 {
	fields := make([]uint32, 16)
	SetUint64(fields, 0, guid)
	fields[2] = 65
	fields[4] = math.Float32bits(1)
	SetUint64(fields, 6, caster)
	fields[8] = uint32(dynamicType)
	fields[9] = uint32(spell)
	fields[10] = math.Float32bits(radius)
	fields[11] = math.Float32bits(x)
	fields[12] = math.Float32bits(y)
	fields[13] = math.Float32bits(z)
	fields[14] = math.Float32bits(orientation)
	return fields
}
