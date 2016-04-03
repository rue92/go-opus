package opus

import (
	"errors"
	. "math"
)

type OpusPacket struct {
	tableOfContents byte
	frames          []OpusFrame
	constantBitRate bool
}

type OpusFrame struct {
	data []byte
}

func newOpusPacket() *OpusPacket {
	p := new(OpusPacket)
	p.frames = nil
	return p
}

func (f *OpusPacket) setTableOfContents(data []byte) {
	f.tableOfContents = data[0]
}

func (packet *OpusPacket) codeZeroPacket(data []byte) *OpusPacket {
	p := *packet
	p.tableOfContents = data[0]
	p.frames = make([]OpusFrame, 1)
	p.frames[0].data = append(p.frames[0].data, data[1:len(data)]...)
	return &p
}

func (packet *OpusPacket) codeOnePacket(data []byte) (*OpusPacket, error) {
	p := *packet
	length := len(data)
	if ((length-1)/2)&0x01 == 0x01 {
		// TODO Figure out how to best return errors
		return nil, errors.New("Invalid Packet")
	}
	p.tableOfContents = data[0]
	p.constantBitRate = true
	p.frames = make([]OpusFrame, 2)
	p.frames[0].data = append(p.frames[0].data, data[1:(length-1)/2]...)
	p.frames[1].data = append(p.frames[1].data, data[(length-1):length]...)
	return &p, nil
}

func (packet *OpusPacket) codeTwoPacket(data []byte) (*OpusPacket, error) {
	p := *packet
	length := len(data)
	p.tableOfContents = data[0]
	p.constantBitRate = false
	firstFrame := uint8(data[1])
	p.frames = make([]OpusFrame, 2)
	p.frames[0].data = append(p.frames[0].data, data[1:firstFrame]...)
	p.frames[1].data = append(p.frames[1].data, data[firstFrame:length]...)
	return &p, nil
}

func (packet *OpusPacket) codeThreePacket(data []byte) (*OpusPacket, error) {
	p := *packet
	//	length := len(data)
	p.tableOfContents = data[0]
	p.constantBitRate = isCBR(data[1])
	paddingRemoved = removePadding(data)
	if p.constantBitRate {
		p.frames, _ = codeThreeCBRPacket(paddingRemoved)
	} else {
		p.frames, _ = codeThreeVBRPacket(paddingRemoved)
	}
	return &p, nil
}

// True means CBR, False means VBR
func isCBR(header byte) bool {
	return header&0x40 != 0
}

func removePadding(data []byte) (trimmed []byte, error error) {
	error = nil
	if !doesPaddingExist(data[1]) {
		return data
	}
	var length uint16

	if data[2] != 255 {
		length = len(data) - data[2]
	} else {
		padLength = 254 + uint16(data[3])
		length = len(data) - padLength
	}
	trimmed := data[0:length]
}

func doesPaddingExist(header byte) bool {
	return header&0x80 != 0
}

func codeThreeCBRPacket(data []byte) ([]OpusFrame, error) {
	return nil, nil
}

func codeThreeVBRPacket(data []byte) ([]OpusFrame, error) {
	return nil, nil
}
