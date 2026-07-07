package media

import (
	"encoding/binary"
	"fmt"

	"github.com/yoshino-s/silk-go"
)

// DefaultSilkSampleRate is the default sample rate for Weixin voice messages.
const DefaultSilkSampleRate = 24000

// SilkToWav decodes a SILK audio buffer and wraps the PCM data in a WAV container.
// sampleRate is the output sample rate in Hz; pass 0 or a negative value to use DefaultSilkSampleRate.
//
// The decoded PCM is mono, 16-bit signed little-endian, matching the 2.4.4 silkToWav behavior.
func SilkToWav(silkBuf []byte, sampleRate int) ([]byte, error) {
	if len(silkBuf) == 0 {
		return nil, fmt.Errorf("silk buffer is empty")
	}
	if sampleRate <= 0 {
		sampleRate = DefaultSilkSampleRate
	}

	pcm, err := silk.DecodePCM(silkBuf, silk.DecodeOptions{
		SampleRateHz: int32(sampleRate),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decode silk: %w", err)
	}
	if len(pcm) == 0 {
		return nil, fmt.Errorf("silk decoder returned empty pcm")
	}

	return pcmBytesToWav(pcm, sampleRate), nil
}

// pcmBytesToWav wraps raw pcm_s16le bytes in a WAV container.
// Mono channel, 16-bit signed little-endian.
func pcmBytesToWav(pcm []byte, sampleRate int) []byte {
	pcmBytes := len(pcm)
	totalSize := 44 + pcmBytes
	buf := make([]byte, totalSize)
	offset := 0

	copy(buf[offset:], "RIFF")
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], uint32(totalSize-8))
	offset += 4
	copy(buf[offset:], "WAVE")
	offset += 4

	copy(buf[offset:], "fmt ")
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], 16)
	offset += 4
	binary.LittleEndian.PutUint16(buf[offset:], 1) // PCM format
	offset += 2
	binary.LittleEndian.PutUint16(buf[offset:], 1) // mono
	offset += 2
	binary.LittleEndian.PutUint32(buf[offset:], uint32(sampleRate))
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], uint32(sampleRate*2)) // byte rate (mono 16-bit)
	offset += 4
	binary.LittleEndian.PutUint16(buf[offset:], 2) // block align
	offset += 2
	binary.LittleEndian.PutUint16(buf[offset:], 16) // bits per sample
	offset += 2

	copy(buf[offset:], "data")
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], uint32(pcmBytes))
	offset += 4

	copy(buf[offset:], pcm)
	return buf
}
