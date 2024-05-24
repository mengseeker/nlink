package transform

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MAX_STREAM_NUM        = 1000_000
	STREAM_READ_CHAN_SIZE = 10

	PACK_HEADER_LEN   = 12
	PACK_MAX_LEN      = 1024 * 4
	PACK_MAX_DATA_LEN = PACK_MAX_LEN - PACK_HEADER_LEN
)

type Pack struct {
	Type       uint32
	Stream     uint32
	DataLength uint32
	Data       [PACK_MAX_LEN]byte
}

const (
	PackFlag_Dial = 1 + iota
	PackFlag_Data
	PackFlag_CloseWrite
	PackFlag_Close
)

func (p *Pack) ReadPackFrom(r io.Reader) error {
	_, err := io.ReadFull(r, p.Data[:PACK_HEADER_LEN])
	if err != nil {
		return err
	}
	p.parseHeader()
	if p.DataLength > 0 {
		_, err := io.ReadFull(r, p.Data[PACK_HEADER_LEN:p.DataLength+PACK_HEADER_LEN])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pack) parseHeader() {
	p.Type = binary.BigEndian.Uint32(p.Data[:4])
	p.Stream = binary.BigEndian.Uint32(p.Data[4:8])
	p.DataLength = binary.BigEndian.Uint32(p.Data[8:12])
}

type PackConn struct {
	conn        *tls.Conn
	packPool    sync.Pool
	writeChan   chan *Pack
	lock        sync.RWMutex
	streams     map[uint32]*PackStream
	curStreamID uint32
	streamsNum  uint32
	closed      atomic.Bool
}

func DialPackConn(addr string, tlsConfig *tls.Config) (*PackConn, error) {
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %v", addr, err)
	}
	pc := PackConn{
		conn: conn,
		packPool: sync.Pool{
			New: func() interface{} {
				return new(Pack)
			},
		},
		writeChan:   make(chan *Pack, 1),
		streams:     make(map[uint32]*PackStream, 1000),
		lock:        sync.RWMutex{},
		curStreamID: 0,
		streamsNum:  0,
	}

	go pc.handleWrite()
	go pc.handleRead()

	return &pc, nil
}

func (pc *PackConn) DialStream(remote *Meta) (*PackStream, error) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if pc.streamsNum >= MAX_STREAM_NUM {
		return nil, fmt.Errorf("too many streams")
	}

	for {
		pc.curStreamID++
		if _, ok := pc.streams[pc.curStreamID]; !ok {
			break
		}
	}

	stream := &PackStream{
		ProxyMeta: *remote,
		conn:      pc,
		streamID:  pc.curStreamID,
		readChan:  make(chan *Pack, STREAM_READ_CHAN_SIZE),
	}

	addr := remote.Marshal()
	pack := pc.packPool.Get().(*Pack)
	pack.Type = PackFlag_Dial
	pack.Stream = stream.streamID
	pack.DataLength = uint32(len(addr))
	copy(pack.Data[:], addr)
	pc.writeChan <- pack

	pc.streams[pc.curStreamID] = stream
	pc.streamsNum++

	return stream, nil
}

func (pc *PackConn) CloseStream(streamID uint32) error {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if _, ok := pc.streams[streamID]; !ok {
		return fmt.Errorf("stream not found")
	}

	pack := pc.packPool.Get().(*Pack)
	pack.Type = PackFlag_Close
	pack.Stream = streamID
	pack.DataLength = 0
	pc.writeChan <- pack

	delete(pc.streams, streamID)
	pc.streamsNum--
	return nil
}

func (pc *PackConn) WritePack(pack *Pack) error {
	if pc.closed.Load() {
		return os.ErrClosed
	}
	pc.writeChan <- pack
	return nil
}

func (pc *PackConn) WritePackWithDeadLine(pack *Pack, deadLine <-chan time.Time) error {
	if pc.closed.Load() {
		return os.ErrClosed
	}
	select {
	case pc.writeChan <- pack:
	case <-deadLine:
		return os.ErrDeadlineExceeded
	}
	return nil
}

func (pc *PackConn) handleWrite() {
	for p := range pc.writeChan {
		_, err := pc.conn.Write(p.Data[:p.DataLength+PACK_HEADER_LEN])
		pc.packPool.Put(p)
		if err != nil {
			fmt.Println("error writing:", err)
		}
	}
}

func (pc *PackConn) handleRead() {
	for {
		pack := pc.packPool.Get().(*Pack)
		err := pack.ReadPackFrom(pc.conn)
		if err != nil {
			fmt.Println("error reading:", err)
		}

		func() {
			pc.lock.RLock()
			defer pc.lock.RUnlock()
			stream, ok := pc.streams[pack.Stream]
			if !ok {
				return
			}
			stream.readChan <- pack
		}()
	}
}

func (pc *PackConn) Close() error {
	pc.closed.Store(true)
	pc.lock.Lock()
	defer pc.lock.Unlock()
	close(pc.writeChan)
	pc.conn.Close()
	return nil
}

type packListener struct {
}

func ListenPackConn(addr string, tlsConfig *tls.Config) (*packListener, error) {
	return nil, nil
}

func (pl *packListener) Accept() (*PackStream, error) {
	return nil, nil
}

var _ net.Conn = (*PackStream)(nil)

type PackStream struct {
	ProxyMeta Meta

	conn       *PackConn
	readChan   chan *Pack
	streamID   uint32
	readClosed atomic.Bool
	closed     atomic.Bool

	readDeadline   time.Time
	curReadPack    *Pack
	curReadPackIdx int

	writeDeadline time.Time
}

func (ps *PackStream) Read(p []byte) (int, error) {
	if ps.readClosed.Load() {
		return 0, io.EOF
	}

	if ps.curReadPack == nil {
		deadLine := time.Until(ps.readDeadline)
		var dead <-chan time.Time
		if deadLine > 0 {
			dead = time.After(deadLine)
		}
		select {
		case ps.curReadPack = <-ps.readChan:
		case <-dead:
			return 0, os.ErrDeadlineExceeded
		}
	}

	if ps.curReadPack == nil {
		return 0, io.EOF
	}

	if ps.curReadPack.Type != PackFlag_Data {
		return 0, fmt.Errorf("type %d not supported", ps.curReadPack.Type)
	}

	if ps.curReadPack.DataLength > 0 && ps.curReadPackIdx < int(ps.curReadPack.DataLength+PACK_HEADER_LEN) {
		n := copy(p, ps.curReadPack.Data[ps.curReadPackIdx:ps.curReadPack.DataLength+PACK_HEADER_LEN])
		ps.curReadPackIdx += n
		return n, nil
	}

	ps.conn.packPool.Put(ps.curReadPack)
	ps.curReadPack = nil
	ps.curReadPackIdx = 0

	return ps.Read(p)
}

func (ps *PackStream) Write(p []byte) (int, error) {
	l := len(p)
	deadLine := time.Until(ps.writeDeadline)
	var dead <-chan time.Time
	if deadLine > 0 {
		dead = time.After(deadLine)
	}

	for i := 0; i < l; {
		i += PACK_MAX_DATA_LEN
		if i > l {
			i = l
		}
		pack := ps.conn.packPool.Get().(*Pack)
		pack.Type = PackFlag_Data
		pack.Stream = ps.streamID
		pack.DataLength = uint32(i)
		copy(pack.Data[:], p[:i])

		if err := ps.conn.WritePackWithDeadLine(pack, dead); err != nil {
			return i, err
		}
		p = p[i:]
	}
	return l, nil
}

func (ps *PackStream) Close() error {
	err := ps.conn.CloseStream(ps.streamID)
	close(ps.readChan)
	return err
}

func (ps *PackStream) LocalAddr() net.Addr {
	return nil
}

func (ps *PackStream) RemoteAddr() net.Addr {
	return nil
}

func (ps *PackStream) SetDeadline(t time.Time) error {
	return nil
}

func (ps *PackStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (ps *PackStream) SetWriteDeadline(t time.Time) error {
	return nil
}
