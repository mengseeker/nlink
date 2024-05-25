package transform

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MAX_STREAM_NUM         = 1000_000
	STREAM_READ_CHAN_SIZE  = 10
	STREAM_WRITE_CHAN_SIZE = 100

	PACK_HEADER_LEN   = 12
	PACK_MAX_LEN      = 1024 * 32
	PACK_MAX_DATA_LEN = PACK_MAX_LEN - PACK_HEADER_LEN
)

//go:generate stringer -type=PackType
type PackType uint32

const (
	PackType_Dial PackType = 1 + iota
	PackType_Data
	PackType_CloseWrite
	PackType_Close
	PackType_Disconnect // client to server only
)

type Pack struct {
	packType   PackType
	stream     uint32
	dataLength uint32
	Buf        [PACK_MAX_LEN]byte
}

func (p *Pack) ReadPackFrom(r io.Reader) error {
	// logger.Debugf("reading pack from %v", r.(net.Conn).RemoteAddr())
	p.reset()

	_, err := io.ReadFull(r, p.header())
	if err != nil {
		return err
	}
	// logger.Debugf("read header: %v", p)
	p.parseHeader()
	if p.dataLength > 0 {
		_, err := io.ReadFull(r, p.Data())
		if err != nil {
			return err
		}
	}
	// logger.Debugf("read data: %v", p)
	return nil
}

func (p *Pack) WritePackTo(w io.Writer) (int, error) {
	return w.Write(p.Buf[:PACK_HEADER_LEN+p.dataLength])
}

func (p *Pack) Type() PackType {
	return p.packType
}

func (p *Pack) Len() int {
	return int(p.dataLength)
}

func (p *Pack) Data() []byte {
	return p.Buf[PACK_HEADER_LEN : p.dataLength+PACK_HEADER_LEN]
}

func (p *Pack) Set(t PackType, streamID uint32, data []byte) *Pack {
	p.packType = t
	p.stream = streamID
	p.dataLength = uint32(len(data))
	binary.BigEndian.PutUint32(p.Buf[:4], uint32(p.packType))
	binary.BigEndian.PutUint32(p.Buf[4:8], p.stream)
	binary.BigEndian.PutUint32(p.Buf[8:12], p.dataLength)
	copy(p.Data(), data)
	return p
}

func (p *Pack) String() string {
	return fmt.Sprintf("type: %v, stream: %d, data length: %d", p.packType, p.stream, p.dataLength)
}

func (p *Pack) reset() {
	p.packType = 0
	p.stream = 0
	p.dataLength = 0
}

func (p *Pack) header() []byte {
	return p.Buf[:PACK_HEADER_LEN]
}

func (p *Pack) parseHeader() {
	p.packType = PackType(binary.BigEndian.Uint32(p.Buf[:4]))
	p.stream = binary.BigEndian.Uint32(p.Buf[4:8])
	p.dataLength = binary.BigEndian.Uint32(p.Buf[8:12])
}
