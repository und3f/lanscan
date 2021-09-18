package scanner

import (
	"fmt"
	"net"
)

type RangeIterator interface {
	HasNext() bool
	GetNext() net.IP
}

type Range interface {
	createIterator() RangeIterator
}

type TwoPointsRange struct {
	start, end net.IP
}

func ParseCIDR(cidr string) (*TwoPointsRange, error) {
	var r *TwoPointsRange
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return r, fmt.Errorf("Failed to parse cidr: %s", err)
	}

	r = new(TwoPointsRange)
	r.start = cloneIP(network.IP)
	r.end = cloneIP(network.IP)

	ones, _ := network.Mask.Size()
	firstSetByte := ones / 8

	if firstSetByte < len(r.end) {
		bitsToSet := ones % 8
		var bits byte
		for bitsCounter, i := byte(0x100>>1), 0; i < bitsToSet; i, bitsCounter = i+1, bitsCounter>>1 {
			bits |= bitsCounter
		}

		r.end[firstSetByte] |= (0xFF ^ bits)

		for i := firstSetByte + 1; i < len(r.end); i++ {
			r.end[i] = 0xFF
		}
	}

	return r, nil
}

func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		//only add to the next byte if we overflowed
		if ip[i] != 0 {
			break
		}
	}
}

func cloneIP(ip net.IP) net.IP {
	cloned := make(net.IP, len(ip))
	copy(cloned, ip)
	return cloned
}

func (r *TwoPointsRange) createIterator() RangeIterator {
	end := cloneIP(r.end)
	incIP(end)
	return &twoPointsRangeIterator{
		end:     end,
		current: cloneIP(r.start),
	}
}

type twoPointsRangeIterator struct {
	current, end net.IP
}

func (iterator *twoPointsRangeIterator) HasNext() bool {
	return !iterator.current.Equal(iterator.end)
}

func (iterator *twoPointsRangeIterator) GetNext() net.IP {
	current := cloneIP(iterator.current)
	incIP(iterator.current)

	return current
}
