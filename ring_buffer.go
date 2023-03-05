package main

import (
	"errors"
)

type RingBuffer struct {
	ready [2]bool
	wpos int32
	rpos int32
	data [32768][256]byte
}

func (s *RingBuffer) Init(index int) {
	s.ready[index] = true
	if index == 0 {
		s.ready[1] = false
		s.wpos = 0
		s.rpos = 0

		for i := 0; i < 32768; i ++ {
			for j := 0; j < 256; j ++ {
				s.data[i][j] = 0
			}
		}
	}
}

func (s *RingBuffer) IsReady() bool {
	return s.ready[0] && s.ready[1]
}

func (s *RingBuffer) Write(d []byte) {
	copy(s.data[s.wpos][0:], d)
	s.wpos = (s.wpos + 1) % 32768
}

func (s *RingBuffer) Read() ([]byte, error) {
	if s.rpos == s.wpos {
		return []byte{}, errors.New("No buffer to read")
	}

	ret := make([]byte, 256)
	copy(ret, s.data[s.rpos][0:])
	s.rpos = (s.rpos + 1) % 32768

	return ret, nil
}

func (s *RingBuffer) IsAvailable() bool {
	return s.rpos != s.wpos
}
