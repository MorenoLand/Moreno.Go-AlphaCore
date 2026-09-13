package packet

import (
	"bytes"
	"encoding/binary"
	"time"
)

type Point struct {
	X, Y, Z float32
}

func EncodeMonsterMove(guid uint64, x, y, z float32, duration uint32, flags uint32, points []Point) ([]byte, error) {
	var data bytes.Buffer
	write := func(value any) { binary.Write(&data, binary.LittleEndian, value) }
	write(guid)
	write(x)
	write(y)
	write(z)
	write(uint32(time.Now().UnixMilli()))
	data.WriteByte(0)
	write(flags)
	write(duration)
	write(uint32(len(points)))
	for _, point := range points {
		write(point.X)
		write(point.Y)
		write(point.Z)
	}
	return Encode(SMSGMonsterMove, data.Bytes())
}
