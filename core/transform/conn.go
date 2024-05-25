package transform

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var (
	packPool = sync.Pool{
		New: func() interface{} {
			return new(Pack)
		},
	}
)

type PackConn struct {
	net.Conn
	curStreamID uint32

	readClosed       bool
	hasReadClosePack bool
	curReadPack      *Pack
	curReadPackIdx   int

	closed atomic.Bool
	isErr  atomic.Bool

	isServer bool
}

func DialPackConn(name, addr string, tlsConfig *tls.Config) (*PackConn, error) {
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %v", addr, err)
	}
	logger.Infof("dial server %s", name)

	pc := PackConn{
		Conn: conn,
	}

	return &pc, nil
}

func AcceptPackConn(conn net.Conn) (*PackConn, error) {
	pc := PackConn{
		Conn:     conn,
		isServer: true,
	}

	return &pc, nil
}

func (pc *PackConn) Accept() (m *Meta, err error) {
	if pc.isErr.Load() {
		return nil, errors.New("connection is broken")
	}

	for {
		pack, err := pc.readPacket()
		if err != nil {
			return nil, fmt.Errorf("read packet error: %v", err)
		}

		if pack.packType != PackType_Dial {
			if pack.packType == PackType_Disconnect {
				packPool.Put(pack)
				return nil, fmt.Errorf("disconnect by client: %s", string(pack.Data()))
			}
			packPool.Put(pack)
			continue
		}

		pc.curStreamID = pack.stream
		pc.curReadPack = nil
		pc.curReadPackIdx = 0
		pc.hasReadClosePack = false
		pc.readClosed = false
		pc.closed.Store(false)

		meta := Meta{}
		meta.Unmarshal(pack.Data())
		return &meta, nil
	}

}

func (pc *PackConn) Bind(m *Meta) error {
	return pc.writePacket(
		packPool.Get().(*Pack).Set(PackType_Dial, pc.curStreamID, m.Marshal()),
	)
}

func (pc *PackConn) CloseWrite() error {
	if pc.closed.Load() {
		return errors.New("connection is closed")
	}

	return pc.writePacket(
		packPool.Get().(*Pack).Set(PackType_CloseWrite, pc.curStreamID, nil),
	)
}

func (pc *PackConn) Close() error {
	if !pc.closed.Load() {
		// logger.Debugf("close connection")
		defer pc.closed.Store(true)
		return pc.writePacket(
			packPool.Get().(*Pack).Set(PackType_Close, pc.curStreamID, nil),
		)
	}
	return nil
}

func (pc *PackConn) Disconnect(reason string) error {
	logger.Warnf("disconnect connection: %s", reason)
	if !pc.isServer {
		pc.writePacket(packPool.Get().(*Pack).Set(PackType_Disconnect, pc.curStreamID, []byte(reason)))
	}

	return pc.Conn.Close()
}

// for client reset the conn
func (pc *PackConn) Reset() error {
	if pc.isErr.Load() {
		return errors.New("connection is broken")
	}

	logger.Debugf("reset connection")

	// try read old packet to clear the buffer
	if !pc.hasReadClosePack {
		for {
			p, err := pc.readPacket()
			if err != nil {
				break
			}
			if p.stream != pc.curStreamID {
				packPool.Put(p)
				return fmt.Errorf("unexpected stream id: %d", p.stream)
			}
			if p.packType == PackType_Close {
				packPool.Put(p)
				break
			}
		}
	}

	logger.Debugf("reset connection done")

	pc.closed.Store(false)
	pc.curStreamID++
	pc.hasReadClosePack = false
	pc.readClosed = false
	pc.curReadPack = nil
	pc.curReadPackIdx = 0

	return nil
}

func (ps *PackConn) Read(p []byte) (int, error) {
	if ps.readClosed {
		return 0, io.EOF
	}

	var err error
	for {
		if ps.curReadPack == nil {
			ps.curReadPack, err = ps.readPacket()
			if err != nil {
				if err == io.EOF {
					return 0, errors.New("connection reset by peer")
				}
				return 0, err
			}

			ps.curReadPackIdx = 0
			if ps.curStreamID != ps.curReadPack.stream {
				return 0, fmt.Errorf("unexpected stream id: %d", ps.curReadPack.stream)
			}

			break
		}

	}

	if ps.curReadPack == nil {
		return 0, io.EOF
	}

	switch ps.curReadPack.packType {
	case PackType_CloseWrite:
		ps.readClosed = true
		return 0, io.EOF
	case PackType_Close:
		ps.readClosed = true
		ps.hasReadClosePack = true
		return 0, io.EOF
	case PackType_Disconnect:
		return 0, fmt.Errorf("unexpected disconnect: %s", string(ps.curReadPack.Data()))
	case PackType_Data:
	default:
		return 0, fmt.Errorf("type %d not supported", ps.curReadPack.Type())
	}

	n := copy(p, ps.curReadPack.Data()[ps.curReadPackIdx:])
	ps.curReadPackIdx += n
	if ps.curReadPackIdx == ps.curReadPack.Len() {
		packPool.Put(ps.curReadPack)
		ps.curReadPack = nil
	}
	return n, nil
}

func (pc *PackConn) Write(p []byte) (int, error) {
	l := len(p)
	cp := p
	for i := 0; i < l; {
		cp = p[i:min(i+PACK_MAX_DATA_LEN, l)]
		i += PACK_MAX_DATA_LEN

		if err := pc.writePacket(
			packPool.Get().(*Pack).Set(PackType_Data, pc.curStreamID, cp),
		); err != nil {
			return 0, err
		}
	}
	return l, nil
}

func (pc *PackConn) writePacket(p *Pack) error {
	// logger.Debugf("write packet, %s", p)
	if pc.closed.Load() {
		return errors.New("connection is closed")
	}
	_, err := p.WritePackTo(pc.Conn)
	if err != nil {
		pc.isErr.Store(true)
		packPool.Put(p)

		logger.Debugf("write packet done, %s, err: %v", p, err)
		return fmt.Errorf("write packet error: %v", err)
	}
	logger.Debugf("write packet done, %s", p)
	return nil
}

func (pc *PackConn) readPacket() (*Pack, error) {
	// logger.Info("read packet")
	p := packPool.Get().(*Pack)
	if err := p.ReadPackFrom(pc.Conn); err != nil {
		packPool.Put(p)
		pc.isErr.Store(true)

		logger.Debugf("read packet done, %s, err: %v", p, err)
		return nil, err
	}

	logger.Debugf("read packet done, %s", p)
	return p, nil
}
