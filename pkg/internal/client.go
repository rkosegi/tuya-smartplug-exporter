/*
Copyright 2026 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/rkosegi/tuya-proto/proto"
)

type Client interface {
	io.Closer
	Read(dest any) error
	Send(cmd proto.CmdIdType, obj any) error
	Connect() error
	IsConnected() bool
	Stats() ProtoStats
}

type clientImpl struct {
	l           *slog.Logger
	to          time.Duration
	rto         time.Duration
	wto         time.Duration
	conn        net.Conn
	key         []byte
	origKey     []byte
	mb          proto.MessageBuilder34
	seqNo       atomic.Uint32
	ver         proto.Version
	clientNonce []byte
	deviceNonce []byte
	host        string
	port        string
	stats       ProtoStats
}

func (c *clientImpl) afterConnect() error {
	var (
		pkt *proto.Packet
		err error
	)
	if c.ver == proto.Version34 {
		c.key = c.origKey
		c.clientNonce = []byte("0123456789abcdef")
		if pkt, err = c.mb.SessKeyNegStart(c.key, c.clientNonce, c.seqNo.Add(1)); err != nil {
			return err
		}
		c.l.Debug("session negotiation step1")
		if err = c.sendPacket(pkt); err != nil {
			return err
		}
		c.l.Debug("session negotiation step2")
		if err = c.readPacket(pkt); err != nil {
			return err
		}
		c.l.Debug("session negotiation step3")
		c.deviceNonce = pkt.DecryptedPayload[:16]
		if pkt, err = c.mb.SessKeyNegFinish(c.key, c.deviceNonce, c.seqNo.Add(1)); err != nil {
			return err
		}
		if err = c.sendPacket(pkt); err != nil {
			return err
		}
		if c.key, err = c.mb.MakeSessionKey(c.clientNonce, c.deviceNonce, c.key); err != nil {
			return err
		}
	}
	return nil
}

func (c *clientImpl) Send(cmd proto.CmdIdType, obj any) error {
	pkt := &proto.Packet{Version: c.ver}
	pkt.SeqNo = c.seqNo.Add(1)
	pkt.CmdId = cmd
	if str, ok := obj.(string); ok {
		pkt.DecryptedPayload = []byte(str)
	} else {
		pkt.SetJsonPayload(obj)
	}
	_, err := pkt.Encode(c.key)
	if err != nil {
		return err
	}
	return c.sendPacket(pkt)
}

func (c *clientImpl) Read(dest any) (err error) {
	pkt := proto.Packet{Version: c.ver}
	if err = c.readPacket(&pkt); err != nil {
		return err
	}
	c.l.Debug("payload decoded", "payload", string(pkt.DecryptedPayload))
	return pkt.GetJsonPayload(dest)
}

func (c *clientImpl) Close() error {
	defer func() {
		c.conn = nil
	}()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *clientImpl) sendPacket(pkt *proto.Packet) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.wto))
	_, err := c.conn.Write(pkt.Encoded())
	c.stats.SentPkts++
	if err != nil {
		c.stats.SentErrs++
	}
	return err
}

func (c *clientImpl) readPacket(pkt *proto.Packet) error {
	_ = c.conn.SetReadDeadline(time.Now().Add(c.rto))
	buf := make([]byte, 4096)
	_, err := c.conn.Read(buf)
	c.stats.ReadPkts++
	if err != nil {
		c.stats.ReadErrs++
		return err
	}
	return pkt.Decode(buf, c.key)
}

type Opt func(*clientImpl)

func WithTimeout(t time.Duration) Opt {
	return func(c *clientImpl) {
		c.to = t
	}
}

func WithReadTimeout(t time.Duration) Opt {
	return func(c *clientImpl) {
		c.rto = t
	}
}

func WithWriteTimeout(t time.Duration) Opt {
	return func(c *clientImpl) {
		c.wto = t
	}
}

func WithLogger(l *slog.Logger) Opt {
	return func(c *clientImpl) {
		c.l = l
	}
}

func (c *clientImpl) IsConnected() bool {
	return c.conn != nil // TODO
}

func NewClient(ver proto.Version, addr string, key []byte, opts ...Opt) Client {
	c := &clientImpl{
		ver:     ver,
		key:     key,
		origKey: key,
		mb:      proto.NewBuilder34(),
	}

	for _, opt := range append([]Opt{
		WithLogger(slog.Default()),
		WithTimeout(time.Second * 10),
		WithReadTimeout(time.Second * 10),
		WithWriteTimeout(time.Second * 10),
	}, opts...) {
		opt(c)
	}
	var err error
	if c.host, c.port, err = net.SplitHostPort(addr); err != nil {
		c.host = addr
		c.port = "6668"
	}

	return c
}

func (c *clientImpl) Connect() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.host, c.port), c.to)
	if err != nil {
		return err
	}
	c.conn = conn
	return c.afterConnect()
}

func (c *clientImpl) Stats() ProtoStats {
	return c.stats
}
