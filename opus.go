package opus

import (
	. "math"
	"errors"
)

type OpusPacket struct {
	tableOfContents byte
	frames []OpusFrame
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
	if ((length-1) / 2) & 0x01 == 0x01 {
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

func getBitrateMode(header byte) bool {
	return header & 0x40 != 0
}

func (packet *OpusPacket) codeThreePacket(data []byte) (*OpusPacket, error) {
	p := *packet
//	length := len(data)
	p.tableOfContents = data[0]
	p.constantBitRate = getBitrateMode(data[1])
	if p.constantBitRate {
		p.frames, _ = codeThreeCBRPacket(data)
	} else {
		p.frames, _ = codeThreeVBRPacket(data)
	}
	return &p, nil
}

func codeThreeCBRPacket(data []byte) ([]OpusFrame, error) {
//	padLength := uint16(data[2]) << 8 | uint16(data[3])
	return nil, nil
}

func codeThreeVBRPacket(data []byte) ([]OpusFrame, error) {
	return nil, nil
}

func opus_pcm_soft_clip(_x []float32, N, C int, declip_mem []float32) {
	var c int
	var i int
	var x []float32

	if C < 1 || N < 1 || _x == nil || declip_mem == nil {
		return
	}
	/* First thing: saturate everything to +/- 2 which is the highest level our
	non-linearity can handle. At the point where the signal reaches +/-2,
	the derivative will be zero anyway, so this doesn't introduce any
	discontinuity in the derivative. */
	for i = 0; i < N*C; i++ {
		_x[i] = float32(Max(-2.0, Min(2.0, float64(_x[i]))))
	}
	for c = 0; c < C; c++ {
		var a, x0 float32
		var curr int

		x = _x[c:]
		a = declip_mem[c]
		/* Continue applying the non-linearity from the previous frame to avoid any discontinuity. */
		for i = 0; i < N; i++ {
			if x[i*C]*a >= 0 {
				break
			}
			x[i*C] = x[i*C] + a*x[i*C]*x[i*C]
		}

		curr = 0
		x0 = x[0]
		for {
			var start, end int
			var maxval float32
			special := false
			var peak_pos int

			for i = curr; i < N; i++ {
				if x[i*C] > 1 || x[i*C] < -1 {
					break
				}
			}
			if i == N {
				a = 0
				break
			}
			peak_pos = i
			start = i
			end = i
			maxval = float32(Abs(float64(x[i*C])))
			/* Look for first zero crossing before clipping */
			for start > 0 && x[i*C]*x[(start-1)*C] >= 0 {
				start--
			}
			/* Look for first zero crossing after clipping */
			for end < N && x[i*C]*x[end*C] >= 0 {
				/* Look for other peaks until the next zero-crossing. */
				if float32(Abs(float64(x[end*C]))) > maxval {
					maxval = float32(Abs(float64(x[end*C])))
					peak_pos = end
				}
				end++
			}
			/* Detect the special case where we clip before the first zero crossing */
			special = (start == 0 && x[i*C]*x[0] >= 0)

			/* Compute a such that maxval + a*maxvval^2 = 1 */
			a = (maxval - 1) / (maxval * maxval)
			if x[i*C] > 0 {
				a = -a
			}
			/* Apply soft clipping */
			for i = start; i < end; i++ {
				x[i*C] = x[i*C] + a*x[i*C]*x[i*C]
			}

			if special && peak_pos >= 2 {
				/* Add a linear ramp from the first sample to the signal peak.
				This avoids a discontinuity at the beginning of the frame. */
				var delta float32
				offset := x0 - x[0]
				delta = offset / float32(peak_pos)
				for i = curr; i < peak_pos; i++ {
					offset -= delta
					x[i*C] += offset
					//					x[i*C] = MAX16(-1.f, MIN16(1.f, x[i*C]));
				}
			}
			curr = end
			if curr == N {
				break
			}
		}
		declip_mem[c] = a
	}
}

func encode_size(size int32, data []uint8) int32 {
	if size < 252 {
		data[0] = uint8(size)
		return 1
	} else {
		data[0] = uint8(252 + (size & 0x3))
		data[1] = uint8((size - int32(data[0])) >> 2)
		return 2
	}
}

func parse_size(data []uint8, len int32, size *int16) int32 {
	if len < 1 {
		*size = -1
		return -1
	} else if data[0] < 252 {
		*size = int16(data[0])
		return 1
	} else if len < 2 {
		*size = -1
		return -1
	} else {
		*size = int16(4*data[1] + data[0])
		return 2
	}
}

func opus_packet_get_samples_per_frame(data []uint8, Fs int32) int32 {
	var audiosize int32
	if data[0] & 0x80 != 0 {
		audiosize = int32((data[0] >> 3) & 0x3)
		audiosize = (Fs << uint32(audiosize)) / 400
	} else if (data[0] & 0x60) == 0x60 {
		if data[0] & 0x08 != 0 {
			audiosize = Fs / 50
		} else {
			audiosize = Fs / 100
		}
	} else {
		audiosize = int32((data[0] >> 3) & 0x3)
		if audiosize == 3 {
			audiosize = Fs * 60 / 1000
		} else {
			audiosize = (Fs << uint32(audiosize)) / 100
		}
	}
	return audiosize
}

/* Not sure about how to do the *frames[48] thing */

// func opus_packet_parse_impl(data *uint8, len int32, self_delimited int32, out_toc *uint8,
// 	frames *[48]uint8, size *[48]int16, payload_offset *int32, packet_offset *int32) int32 {
	
// 	var i, bytes, count, cbr, framesize, last_size int32
// 	var pad int32 = 0
// 	var ch, toc uint8
// 	data0 := data

// 	if size == nil {
// 		return 0xFF // Replace with OPUS_BAD_ARG
// 	}

// 	framesize = opus_packet_get_samples_per_frame(data, 48000)

// 	cbr = 0
// 	toc = *data
// 	*data = *data + 1
// 	len--
// 	last_size = len
// 	switch (toc & 0x03) {
// 		/* One frame */
// 	case 0:
// 		count = 1;
// 		break
// 		/*Two CBR frames */
// 	case 1:
// 		count = 2
// 		cbr = 1
// 		if !self_delimited {
// 			if len & 0x1 {
// 				return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 			}
// 			last_size = len / 2
// 			size[0] = int16(last_size)
// 		}
// 		break
// 		/* Two VBR frames */
// 	case 2:
// 		count = 2
// 		bytes = parse_size(data, len, size)
// 		len -= bytes
// 		if size[0] < 0 || size[0] > len {
// 			return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 		}
// 		data += bytes
// 		last_size = len - size[0]
// 		break
// 		/* Multiple CBR/VBR frames (from 0 to 120 ms) */
// 	default:
// 		if len < 1 {
// 			return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 		}
// 		/* Number of frames encoded in bits 0 to 5 */
// 		ch = *data
// 		*data = *data + 1
// 		count = ch & 0x3F
// 		if count <= 0 || framesize*count > 5760 {
// 			return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 		}
// 		len--
// 		/* Padding flag is bit 6 */
// 		if ch & 0x40 {
// 			var p int32
// 			for {
// 				var tmp int32
// 				if len <= 0 {
// 					return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 				}
// 				p = *data
// 				*data = *data + 1
// 				len--
// 				if p == 255 { tmp = 254 } else { tmp = p }
// 				len -= tmp
// 				pad += tmp

// 				if p != 255 {
// 					break
// 				}
// 			}
// 		}
// 		if len < 0 {
// 			return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 		}
// 		/* VBR flag is bit 7 */
// 		cbr = !(ch & 0x80)
// 		if !cbr {
// 			/* VBR case */
// 			last_size = len
// 			for i = 0; i < count - 1; i++ {
// 				bytes = parse_size(data, len, size+i)
// 				len -= bytes
// 				if size[i] < 0 || size[i] > len {
// 					return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 				}
// 				data += bytes
// 				last_size -= bytes + size[i]
// 			}
// 			if last_size < 0 {
// 				return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 			}
// 		} else if !self_delimited {
// 			/* CBR case */
// 			last_size = len / count
// 			if last_size * count != len {
// 				return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 			}
// 			for i = 0; i < count - 1; i++ {
// 				size[i] = int16(last_size)
// 			}
// 		}
// 		break
// 	}
// 		/* Self-delimited framing has an extra size for the last frame. */
// 		if self_delimited {
// 			bytes = parse_size(data, len, size + count - 1)
// 			len -= bytes
// 			if size[count - 1] < 0 || size[count - 1] > len {
// 				return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 			}
// 			data += bytes
// 			/* For CBR packets, apply the size to all the frames. */
// 			if cbr {
// 				if size[count - 1] * count > len {
// 					return OPUS_INVALID_PACKET
// 				}
// 				for i = 0; i < count - 1; i++ {
// 					size[i] = size[count - 1]
// 				}
// 			} else if bytes + size[count - 1] > last_size {
// 				return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 			}
// 		} else {
// 			/* Because it's not encoded explicitly, it's possible the 
// size of the last packet (or all the packets, for the CBR case) is larger than
// 1275. Reject them here. */
// 			if last_size > 1275 {
// 				return 0x08 // TODO: Use OPUS_INVALID_PACKET
// 			}
// 			size[count - 1] = int16(last_size)
// 		}

// 		if payload_offset {
// 			*payload_offset = int32(data - data0)
// 		}

// 		for i = 0; i < count; i++ {
// 			if frames {
// 				frames[i] = data
// 			}
// 			data += size[i]
// 		}

// 		if packet_offset {
// 			*packet_offset = pad + int32(data - data0)
// 		}

// 		if out_toc {
// 			*out_toc = toc
// 		}

// 		return count
// }

// func opus_packet_parse(data []uint8, len int32, out_toc *uint8, frames *[48]uint8,
// 	size *[48]uint16, payload_offset *int32) {
// 	return opus_packet_parse_impl(data, len, 0, out_toc, frames, size, payload_offset, nil)
// }
