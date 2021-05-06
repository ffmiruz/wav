// This package implements WAV file encoding and decoding.
package wav

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	// ASCII string "RIFF"
	RIFF_TAG uint32 = 0x52494646
	// ASCII string "WAVE"
	WAVE_TAG uint32 = 0x57415645
	// ASCII string "fmt "
	fmt_TAG uint32 = 0x666d7420
	// ASCII string "data"
	data_TAG uint32 = 0x64617461
)

type File struct {
	Header
	FmtChunk
	DataChunk
}

// WAVE file header/descriptor chunk.
//
// Id is ASCII of "RIFF" or 0x52494646.
//
// Size is the size of entire file excluding Chunk Id and Chunk Size.
// 4 + (8 + FmtChunkSize) + (8 + DataChunkSize). Equivalent to the
// size of the file from this point onwards.
//
// Format is ASCII of "WAVE" or 0x57415645
type Header struct {
	Id     uint32 //  4 bytes big endian
	Size   uint32 //  4 bytes little endian
	Format uint32 //  4 bytes big endian
}

// fmt sub-chunk.
type FmtChunk struct {
	ID          uint32 // 4 bytes big endian
	Size        uint32 // 4 bytes little endian
	AudioFormat uint16 // 2 bytes little endian
	Channel     uint16 // 2 bytes little endian
	// Sample per second
	SampleRate    uint32 // 4 bytes little endian
	ByteRate      uint32 // 4 bytes little endian
	BlockAlign    uint16 // 2 bytes little endian
	BitsPerSample uint16 // 2 bytes little endian
}

// data sub-chunk.
type DataChunk struct {
	ID   uint32 // 4 bytes big endian
	Size uint32 // 4 bytes little endian
	// Data hold slice of Sample for each channel
	Data []Sample
}

// Channel data samples.
type Sample []int

func Unmarshal(r io.Reader) (File, error) {
	file := File{}
	err := file.Header.Unmarshal(r)
	if err != nil {
		return file, err
	}
	err = file.FmtChunk.Unmarshal(r)
	if err != nil {
		return file, err
	}
	err = file.DataChunk.Unmarshal(r)
	if err != nil {
		return file, err
	}
	err = file.ParseData(r)
	if err != nil {
		return file, err
	}
	return file, err

}

func (h *Header) Unmarshal(r io.Reader) error {
	buf := make([]byte, 12)
	_, err := r.Read(buf)
	if err != nil {
		return errors.New("fail to read header")
	}
	h.Id = binary.BigEndian.Uint32(buf[0:4])
	if h.Id != RIFF_TAG {
		return errors.New("not a valid RIFF file")
	}
	// TODO: check for file size
	h.Size = binary.LittleEndian.Uint32(buf[4:8])
	h.Format = binary.BigEndian.Uint32(buf[8:12])
	if h.Format != WAVE_TAG {
		return errors.New("not a valid WAVE file")
	}
	return err
}

func (fm *FmtChunk) Unmarshal(r io.Reader) error {
	buf := make([]byte, 24)
	_, err := r.Read(buf)
	if err != nil {
		return errors.New("fail to read fmt chunk")
	}
	fm.ID = binary.BigEndian.Uint32(buf[0:4])
	if fm.ID != fmt_TAG {
		return errors.New("not a valid fmt chunk")
	}
	fm.Size = binary.LittleEndian.Uint32(buf[4:8])
	fm.AudioFormat = binary.LittleEndian.Uint16(buf[8:10])
	fm.Channel = binary.LittleEndian.Uint16(buf[10:12])
	fm.SampleRate = binary.LittleEndian.Uint32(buf[12:16])
	fm.ByteRate = binary.LittleEndian.Uint32(buf[16:20])
	fm.BlockAlign = binary.LittleEndian.Uint16(buf[20:22])
	fm.BitsPerSample = binary.LittleEndian.Uint16(buf[22:24])
	return err
}

// TODO: sanity check
func (dat *DataChunk) Unmarshal(r io.Reader) error {
	buf := make([]byte, 8)
	_, err := r.Read(buf)
	if err != nil {
		return errors.New("fail to read data chunk header")
	}
	dat.ID = binary.BigEndian.Uint32(buf[0:4])
	if dat.ID != data_TAG {
		return errors.New("not a valid data chunk")
	}
	dat.Size = binary.LittleEndian.Uint32(buf[4:8])
	return err
}

func (file *File) ParseData(r io.Reader) error {
	var err error
	// Number of sample in the data
	sampleCount := int(file.DataChunk.Size) / int(file.FmtChunk.Channel) / int(file.FmtChunk.BitsPerSample/8)
	data := make([]Sample, file.FmtChunk.Channel)

	// Create channel samples slice for each channel
	for ch := range data {
		sample := make([]int, sampleCount)
		data[ch] = sample
	}

	// Enough buffer to fit Uint32 for 32 bits per sample
	buf := make([]byte, 4)
	for sample := 0; sample < sampleCount; sample++ {
		for channel := 0; channel < len(data); channel++ {
			_, err := r.Read(buf)
			if err != nil && err != io.EOF {
				return errors.New("fail to parse sample data")
			}
			switch file.FmtChunk.BitsPerSample {
			case 8:
				data[channel][sample] = int(int8(buf[0]))
			case 16:
				data[channel][sample] = int(int16(binary.LittleEndian.Uint16(buf)))
			case 32:
				data[channel][sample] = int(int32(binary.LittleEndian.Uint32(buf)))
			default:
				return errors.New("unsupported bits per sample")

			}
		}
	}
	file.DataChunk.Data = data
	return err
}

func Marshal(file File) []byte {
	buf := make([]byte, 4+4+file.Header.Size)

	binary.BigEndian.PutUint32(buf[0:4], file.Header.Id)
	binary.LittleEndian.PutUint32(buf[4:8], file.Header.Size)
	binary.BigEndian.PutUint32(buf[8:12], file.Header.Format)

	binary.BigEndian.PutUint32(buf[12:16], file.FmtChunk.ID)
	binary.LittleEndian.PutUint32(buf[16:20], file.FmtChunk.Size)
	binary.LittleEndian.PutUint16(buf[20:22], file.FmtChunk.AudioFormat)
	binary.LittleEndian.PutUint16(buf[22:24], file.FmtChunk.Channel)
	binary.LittleEndian.PutUint32(buf[24:28], file.FmtChunk.SampleRate)
	binary.LittleEndian.PutUint32(buf[28:32], file.FmtChunk.ByteRate)
	binary.LittleEndian.PutUint16(buf[32:34], file.FmtChunk.BlockAlign)
	binary.LittleEndian.PutUint16(buf[34:36], file.FmtChunk.BitsPerSample)

	binary.BigEndian.PutUint32(buf[36:40], file.DataChunk.ID)
	binary.LittleEndian.PutUint32(buf[40:44], file.DataChunk.Size)

	file.MarshalData(buf[44:])

	return buf

}

func (file File) MarshalData(buf []byte) {

	sampleCount := len(file.DataChunk.Data[0])
	data := file.DataChunk.Data

	// pos is current position in buf.
	for sample, pos := 0, 0; sample < sampleCount; sample++ {
		for channel := range data {
			binary.LittleEndian.PutUint16(buf[pos:pos+2], uint16(int(data[channel][sample])))
			pos = pos + 2
		}
	}
}
