package internal

import (
	"context"
	"net"
	"time"

	"github.com/Dreamacro/clash/constant"
)

var _ net.Conn = (*wrapPacketConnIntoConn)(nil)

type wrapPacketConnIntoConn struct {
	addr net.Addr
	constant.PacketConn
}

func newWrapPacketConnIntoConn(pc constant.PacketConn) net.Conn {
	addr, _ := net.ResolveUDPAddr("udp", "1.1.1.1:53")
	return &wrapPacketConnIntoConn{addr, pc}
}

func (w *wrapPacketConnIntoConn) Read(b []byte) (n int, err error) {
	n, _, err = w.ReadFrom(b)
	return
}

func (w *wrapPacketConnIntoConn) Write(b []byte) (n int, err error) {
	return w.WriteTo(b, w.addr)
}

func (w *wrapPacketConnIntoConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func DnsResovleOverProxy(domain string, proxy constant.Proxy) ([]net.IP, error) {
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			metadata := constant.Metadata{
				DstIP:   nil,
				Host:    "1.1.1.1",
				DstPort: "53",
			}
			pc, err := proxy.DialUDP(&metadata)
			return newWrapPacketConnIntoConn(pc), err
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return resolver.LookupIP(ctx, "ip", domain)
}

func IsUDPForwardAvailable(proxy constant.Proxy) bool {
	_, err := DnsResovleOverProxy("www.google.com", proxy)
	return err == nil
}
