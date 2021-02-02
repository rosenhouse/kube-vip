package vip

import (
	"fmt"
	"net"

	"github.com/mdlayher/ndp"

	log "github.com/sirupsen/logrus"
)

type NdpResponder struct {
	intf         string
	hardwareAddr net.HardwareAddr
	conn         *ndp.Conn
}

func NewNDPResponder(ifi *net.Interface) (*NdpResponder, error) {
	// Use link-local address as the source IPv6 address for NDP communications.
	conn, _, err := ndp.Dial(ifi, ndp.LinkLocal)
	if err != nil {
		return nil, fmt.Errorf("creating NDP responder for %q: %s", ifi.Name, err)
	}

	ret := &NdpResponder{
		intf:         ifi.Name,
		hardwareAddr: ifi.HardwareAddr,
		conn:         conn,
	}
	return ret, nil
}

func (n *NdpResponder) Close() error {
	return n.conn.Close()
}

func (n *NdpResponder) Gratuitous(ip net.IP) error {
	return n.advertise(net.IPv6linklocalallnodes, ip, true)
}

func (n *NdpResponder) advertise(dst, target net.IP, gratuitous bool) error {
	m := &ndp.NeighborAdvertisement{
		Solicited:     !gratuitous,
		Override:      gratuitous, // Should clients replace existing cache entries
		TargetAddress: target,
		Options: []ndp.Option{
			&ndp.LinkLayerAddress{
				Direction: ndp.Target,
				Addr:      n.hardwareAddr,
			},
		},
	}
	log.Infof("ndp: %v", m)
	return n.conn.WriteTo(m, nil, dst)
}
