package frameencoder

import (
	"bytes"
	"encoding/binary"
)

// sizeof(FrameHeader) = 2 + 2 + 2 + 1 + 1 + 256 + 16 + 4 + 4 + 2 + 2 + 2 + 276*2 + 128 = 974
type FrameHeader struct {
	Magic                uint16
	ImgWidth             uint16
	ImgHeight            uint16
	TileWidth            uint8
	TileHeight           uint8
	P8RunBits            []uint8 // 256 elements
	P4RunBits            []uint8 // 16 elements
	P2RunBits            []uint8 // 4 elements
	SelectorRunBits      []uint8 // 4 elements
	TileChangeRunBits    []uint8 // 2 elements
	LineChangeRunBits    []uint8 // 2 elements
	Reserved0            uint16
	Palette              []uint16 // 276 elements (4 colors for P2, 16 colors for P4, 256 colors for P8)
	LineChangeStreamData [128]uint8
}

func NewFrameHeader() *FrameHeader {
	return &FrameHeader{
		Magic: 0x4645, // 'FE' in ASCII
	}
}

func (fh *FrameHeader) SetImgDimensions(width, height uint16) {
	fh.ImgWidth = width
	fh.ImgHeight = height
}

func (fh *FrameHeader) SetTileDimensions(tileWidth, tileHeight uint8) {
	fh.TileWidth = tileWidth
	fh.TileHeight = tileHeight
}

func (fh *FrameHeader) SetLineChange(lineChangeBits []uint8) {
	copy(fh.LineChangeRunBits, lineChangeBits)
}

func (fh *FrameHeader) SetPalette(palette []uint16) {
	fh.Palette = palette
}

func (fh *FrameHeader) WriteBinary(dst bytes.Buffer) {

	// Write the frame header to the provided buffer in binary format
	binary.Write(&dst, binary.LittleEndian, fh.Magic)
	binary.Write(&dst, binary.LittleEndian, fh.ImgWidth)
	binary.Write(&dst, binary.LittleEndian, fh.ImgHeight)
	binary.Write(&dst, binary.LittleEndian, fh.TileWidth)
	binary.Write(&dst, binary.LittleEndian, fh.TileHeight)

	binary.Write(&dst, binary.LittleEndian, fh.P8RunBits)
	binary.Write(&dst, binary.LittleEndian, fh.P4RunBits)
	binary.Write(&dst, binary.LittleEndian, fh.P2RunBits)
	binary.Write(&dst, binary.LittleEndian, fh.SelectorRunBits)
	binary.Write(&dst, binary.LittleEndian, fh.TileChangeRunBits)
	binary.Write(&dst, binary.LittleEndian, fh.LineChangeRunBits)
	binary.Write(&dst, binary.LittleEndian, fh.Reserved0)
	binary.Write(&dst, binary.LittleEndian, fh.Palette)

	binary.Write(&dst, binary.LittleEndian, fh.LineChangeStreamData)
}
