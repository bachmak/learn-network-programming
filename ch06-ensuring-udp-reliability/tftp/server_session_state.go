package tftp

import "time"

type sessionState struct {
	clientName string
	block      uint16
	retries    uint8
	timeout    time.Duration
}

func (s sessionState) incrementBlock() sessionState {
	return sessionState{
		clientName: s.clientName,
		block:      s.block + 1,
		retries:    s.retries,
		timeout:    s.timeout,
	}
}

func (s sessionState) decrementRetries() sessionState {
	return sessionState{
		clientName: s.clientName,
		block:      s.block,
		retries:    s.retries - 1,
		timeout:    s.timeout,
	}
}
