// File: internal/server/pty.go
// Project: Terminal Velocity
// Description: Parse SSH pty-req and window-change request payloads (RFC 4254 §6.2, §6.7)
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-22

package server

import (
	"encoding/binary"
	"errors"
)

// ptySize carries character dimensions from SSH pty-req / window-change requests.
// Pixel dimensions are intentionally dropped — the TUI only cares about character cells.
type ptySize struct {
	cols uint32
	rows uint32
}

// fallbackPTYSize is what we report to BubbleTea when the client never issued a
// pty-req or the payload was malformed. 80x24 matches VT100 tradition and keeps
// the TUI usable until the client sends a real size.
var fallbackPTYSize = ptySize{cols: 80, rows: 24}

// ErrPTYPayloadTooShort indicates the request payload was truncated relative to
// the RFC 4254 expected layout.
var ErrPTYPayloadTooShort = errors.New("pty request payload too short")

// parsePTYReq decodes an SSH "pty-req" channel request payload per RFC 4254 §6.2.
//
// Layout (big-endian):
//
//	string  TERM       (uint32 length + bytes)
//	uint32  width chars
//	uint32  height rows
//	uint32  width px   (ignored)
//	uint32  height px  (ignored)
//	string  modes      (ignored)
//
// Returns the character dimensions and (when present) TERM string.
func parsePTYReq(payload []byte) (ptySize, string, error) {
	term, rest, err := readSSHString(payload)
	if err != nil {
		return ptySize{}, "", err
	}
	if len(rest) < 16 {
		return ptySize{}, "", ErrPTYPayloadTooShort
	}
	cols := binary.BigEndian.Uint32(rest[0:4])
	rows := binary.BigEndian.Uint32(rest[4:8])
	// pixel-width (rest[8:12]) and pixel-height (rest[12:16]) ignored
	return ptySize{cols: cols, rows: rows}, string(term), nil
}

// parseWindowChange decodes an SSH "window-change" channel request payload per
// RFC 4254 §6.7.
//
// Layout (big-endian):
//
//	uint32  width chars
//	uint32  height rows
//	uint32  width px   (ignored)
//	uint32  height px  (ignored)
func parseWindowChange(payload []byte) (ptySize, error) {
	if len(payload) < 16 {
		return ptySize{}, ErrPTYPayloadTooShort
	}
	cols := binary.BigEndian.Uint32(payload[0:4])
	rows := binary.BigEndian.Uint32(payload[4:8])
	return ptySize{cols: cols, rows: rows}, nil
}

// readSSHString reads an SSH wire-format string (uint32 length + bytes) from the
// front of buf, returning the string body and the remainder of the buffer.
func readSSHString(buf []byte) ([]byte, []byte, error) {
	if len(buf) < 4 {
		return nil, nil, ErrPTYPayloadTooShort
	}
	length := binary.BigEndian.Uint32(buf[0:4])
	if uint32(len(buf)-4) < length {
		return nil, nil, ErrPTYPayloadTooShort
	}
	return buf[4 : 4+length], buf[4+length:], nil
}

// valid reports whether both dimensions are non-zero. Zero-valued sizes come
// from clients that didn't ask for a PTY; we fall back to defaults in that case.
func (p ptySize) valid() bool {
	return p.cols > 0 && p.rows > 0
}
