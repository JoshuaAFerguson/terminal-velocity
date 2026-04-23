// File: internal/server/pty_test.go
// Project: Terminal Velocity
// Description: Tests for SSH pty-req and window-change payload parsing
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2026-04-22

package server

import (
	"encoding/binary"
	"errors"
	"testing"
)

func buildPTYReqPayload(term string, cols, rows, pxW, pxH uint32, modes string) []byte {
	buf := make([]byte, 4+len(term)+16+4+len(modes))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(term)))
	copy(buf[4:4+len(term)], term)
	off := 4 + len(term)
	binary.BigEndian.PutUint32(buf[off:off+4], cols)
	binary.BigEndian.PutUint32(buf[off+4:off+8], rows)
	binary.BigEndian.PutUint32(buf[off+8:off+12], pxW)
	binary.BigEndian.PutUint32(buf[off+12:off+16], pxH)
	off += 16
	binary.BigEndian.PutUint32(buf[off:off+4], uint32(len(modes)))
	copy(buf[off+4:off+4+len(modes)], modes)
	return buf
}

func buildWindowChangePayload(cols, rows, pxW, pxH uint32) []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(buf[0:4], cols)
	binary.BigEndian.PutUint32(buf[4:8], rows)
	binary.BigEndian.PutUint32(buf[8:12], pxW)
	binary.BigEndian.PutUint32(buf[12:16], pxH)
	return buf
}

func TestParsePTYReq(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		wantSize ptySize
		wantTerm string
		wantErr  error
	}{
		{
			name:     "xterm 120x40",
			payload:  buildPTYReqPayload("xterm-256color", 120, 40, 960, 640, ""),
			wantSize: ptySize{cols: 120, rows: 40},
			wantTerm: "xterm-256color",
		},
		{
			name:     "tiny terminal",
			payload:  buildPTYReqPayload("vt100", 40, 12, 0, 0, "\x00"),
			wantSize: ptySize{cols: 40, rows: 12},
			wantTerm: "vt100",
		},
		{
			name:     "empty TERM is allowed",
			payload:  buildPTYReqPayload("", 80, 24, 0, 0, ""),
			wantSize: ptySize{cols: 80, rows: 24},
			wantTerm: "",
		},
		{
			name:     "zero size (client didn't want a real PTY)",
			payload:  buildPTYReqPayload("dumb", 0, 0, 0, 0, ""),
			wantSize: ptySize{cols: 0, rows: 0},
			wantTerm: "dumb",
		},
		{
			name:    "truncated before TERM length",
			payload: []byte{0x00, 0x00},
			wantErr: ErrPTYPayloadTooShort,
		},
		{
			name:    "TERM length longer than payload",
			payload: []byte{0x00, 0x00, 0x00, 0xff, 'a', 'b'},
			wantErr: ErrPTYPayloadTooShort,
		},
		{
			name:    "missing dimensions after TERM",
			payload: []byte{0x00, 0x00, 0x00, 0x05, 'x', 't', 'e', 'r', 'm'},
			wantErr: ErrPTYPayloadTooShort,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			size, term, err := parsePTYReq(tc.payload)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want err %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if size != tc.wantSize {
				t.Errorf("size: want %+v, got %+v", tc.wantSize, size)
			}
			if term != tc.wantTerm {
				t.Errorf("term: want %q, got %q", tc.wantTerm, term)
			}
		})
	}
}

func TestParseWindowChange(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		wantSize ptySize
		wantErr  error
	}{
		{
			name:     "resize to 200x60",
			payload:  buildWindowChangePayload(200, 60, 1600, 960),
			wantSize: ptySize{cols: 200, rows: 60},
		},
		{
			name:     "minimal 80x24",
			payload:  buildWindowChangePayload(80, 24, 0, 0),
			wantSize: ptySize{cols: 80, rows: 24},
		},
		{
			name:    "too short",
			payload: []byte{0x00, 0x00, 0x00, 0x50},
			wantErr: ErrPTYPayloadTooShort,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			size, err := parseWindowChange(tc.payload)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want err %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if size != tc.wantSize {
				t.Errorf("size: want %+v, got %+v", tc.wantSize, size)
			}
		})
	}
}

func TestPTYSizeValid(t *testing.T) {
	tests := []struct {
		name string
		in   ptySize
		want bool
	}{
		{"zero", ptySize{}, false},
		{"zero cols", ptySize{cols: 0, rows: 24}, false},
		{"zero rows", ptySize{cols: 80, rows: 0}, false},
		{"both set", ptySize{cols: 80, rows: 24}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.valid(); got != tc.want {
				t.Errorf("valid(): want %v, got %v", tc.want, got)
			}
		})
	}
}
