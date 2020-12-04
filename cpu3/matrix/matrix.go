package matrix

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/byuoitav/connpool"
	"go.uber.org/zap"
)

const (
	asciiCarriageReturn = 0x0d
	asciiPromptEnd      = 0x3e
)

type Matrix struct {
	pool *connpool.Pool
	log  *zap.Logger

	// OutputSlotStart is the number that this switcher starts its output slot
	// numbers on when you run `dumpdmrouteinfo`. For example,
	// if the first output card is slot 33, this number should be 33.
	OutputSlotStart int

	// SetRouteOutputStart this is some ...randomly generated number that crestron uses
	// when you run setavroute. You can find the number by running `cards` - it should be output ID that
	// matches the card at OutputSlotStart (ie, running `setavroute <input> SetRouteOutputStart` should affect
	// the output at OutputSlotStart).
	SetRouteOutputStart int
}

func New(addr string, opts ...Option) Matrix {
	options := &options{
		ttl:   60 * time.Second,
		delay: 500 * time.Millisecond,
		log:   zap.NewNop(),
	}

	for _, o := range opts {
		o.apply(options)
	}

	return Matrix{
		pool: &connpool.Pool{
			TTL:   options.ttl,
			Delay: options.delay,
			NewConnection: func(ctx context.Context) (net.Conn, error) {
				dialer := net.Dialer{}

				conn, err := dialer.DialContext(ctx, "tcp", addr+":23")
				if err != nil {
					return nil, err
				}

				deadline, ok := ctx.Deadline()
				if !ok {
					deadline = time.Now().Add(5 * time.Second)
				}

				conn.SetDeadline(deadline)

				// read the until the prompt
				buf := make([]byte, 64)
				for !bytes.Contains(buf, []byte{asciiPromptEnd}) {
					_, err := conn.Read(buf)
					if err != nil {
						conn.Close()
						return nil, fmt.Errorf("unable to read new connection prompt: %w", err)
					}
				}

				return conn, nil
			},
		},
		log: options.log,
	}
}

func (m *Matrix) sendCommand(ctx context.Context, cmd []byte) ([]byte, error) {
	var resp []byte

	err := m.pool.Do(ctx, func(conn connpool.Conn) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			deadline = time.Now().Add(10 * time.Second)
		}

		conn.SetDeadline(deadline)

		n, err := conn.Write(cmd)
		switch {
		case err != nil:
			return fmt.Errorf("unable to write command: %w", err)
		case n != len(cmd):
			return fmt.Errorf("unable to write command: wrote %v/%v bytes", n, len(cmd))
		}

		r, err := conn.ReadUntil(asciiPromptEnd, deadline)
		if err != nil {
			return fmt.Errorf("unable to read response: %w", err)
		}

		resp = bytes.TrimSpace(r)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
