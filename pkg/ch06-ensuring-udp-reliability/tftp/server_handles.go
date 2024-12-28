package tftp

import (
	"fmt"
)

func handleMsg(msg []byte, state sessionState) error {
	ackPacket, err := tryHandleAckMsg(msg)
	if err == nil {
		if uint16(ackPacket) == state.block {
			return nil
		} else {
			return fmt.Errorf(
				"[%s] unexpected block acked: %d instead of %d",
				state.clientName,
				uint16(ackPacket),
				state.block,
			)
		}
	}
	errPacket, err := tryHandleErrMsg(msg)
	if err == nil {
		return fmt.Errorf(
			"[%s] received error message: %s",
			state.clientName,
			errPacket.Message,
		)
	}
	return fmt.Errorf("[%s] bad packet", state.clientName)
}

func tryHandleAckMsg(msg []byte) (Ack, error) {
	var ackPacket Ack
	err := ackPacket.UnmarshalBinary(msg)
	if err == nil {
		return ackPacket, nil
	}
	return 0, err
}

func tryHandleErrMsg(msg []byte) (Err, error) {
	var errPacket Err
	err := errPacket.UnmarshalBinary(msg)
	if err == nil {
		return errPacket, nil
	}
	return Err{}, err
}
