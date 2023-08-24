package fontcompress

import (
	"errors"
	"io"
	"os"
)

const (
	// TTF
	TTF_MAGIC uint32 = 0x00010000
	// OTF
	OTF_MAGIC uint32 = 0x74727565
)

type TTFTable interface{}

type TTFTableInfo struct {
	Tag      uint32 // 4-byte identifier;
	CheckSum uint32 // CheckSum for this table
	Offset   uint32 // Offset from beginning of TrueType font file
	Length   uint32 // Length of this table

	Table TTFTable
}

type CmapSubHeader struct {
	// First character code covered
	FirstCode uint16
	// Last character code covered
	EntryCount uint16
	// Glyph index array
	IdDelta uint16
	// Number of idRangeOffsets that follow
	IdRangeOffset uint16
}

type CmapSubTableFormatMixGroup struct {
	/**
	UInt32	startCharCode	First character code in this group; note that if this group is for one or more 16-bit character codes (which is determined from the is32 array), this 32-bit value will have the high 16-bits set to zero
	UInt32	endCharCode	Last character code in this group; same condition as listed above for the startCharCode
	UInt32	startGlyphCode	Glyph index corresponding to the starting character code
	*/
	StartCharCode  uint32
	EndCharCode    uint32
	StartGlyphCode uint32
}
type CmapSubTable struct {
	// platform:
	// 0 - Unicode
	// 1 - Macintosh
	// 2 - reserved
	// 3 - Microsoft
	PlatformID uint16 // Platform ID
	// encoding:
	// 0 - Unicode 1.0 semantics
	// 1 - Unicode 1.1 semantics
	// 2 - ISO/IEC 10646 semantics (deprecated)
	// 3 - Unicode 2.0 and onwards semantics, Unicode BMP only (cmap subtable formats 0, 4, 6, 10, 12)
	// 4 - Unicode 2.0 and onwards semantics, Unicode full repertoire (cmap subtable formats 0, 4, 6, 10, 12, 13)
	// 5 - Unicode Variation Sequences (cmap subtable format 14)
	// 6 - Last Resort Font semantics
	EncodingID uint16 // Platform-specific encoding ID
	SubOffset  uint32 // Byte offset from beginning of table to the subtable for this encoding

	// format:
	// 0 - Byte encoding table
	// 2 - High-byte mapping through table
	// 4 - Segment mapping to delta values
	// 6 - Trimmed table mapping
	Format uint16 // Format number is set to 0
	// length:
	// This is the length in bytes of the subtable.
	// If it is not exactly the size needed to contain the subtable,
	// then the subtable should be treated as if this field were set to 0.
	// See the note below for more information.
	Length uint16 // Length in bytes of the subtable (including this header)
	// language:
	// For requirements on use of the language field,
	// see “Use of the language field in ‘cmap’ subtables” in this document.
	// For information on the language field for Windows platforms,
	// see the “EncodingID” field description in the “name” table chapter.
	Language uint16 // Language code for this encoding subtable, or zero if language-independent
	// An array that maps character codes to glyph index values
	GlyphIndexArray []uint8

	// format 2
	// Array that maps high bytes to subHeaders: value is index * 8
	SubHeaderKeys []uint16
	// This is an array of subHeaders.
	SubHeaders []CmapSubHeader
	// This is an array of glyphIndexArray elements:
	GlyphIndexArray16 []uint16

	// format 4
	/**
	UInt16	segCountX2	2 * segCount
	UInt16	searchRange	2 * (2**FLOOR(log2(segCount)))
	UInt16	entrySelector	log2(searchRange/2)
	UInt16	rangeShift	(2 * segCount) - searchRange
	UInt16	endCode[segCount]	Ending character code for each segment, last = 0xFFFF.
	UInt16	reservedPad	This value should be zero
	UInt16	startCode[segCount]	Starting character code for each segment
	UInt16	idDelta[segCount]	Delta for all character codes in segment
	UInt16	idRangeOffset[segCount]	Offset in bytes to glyph indexArray, or 0
	*/
	SegCountX2 uint16
	// 2 * (2**FLOOR(log2(segCount)))
	SearchRange uint16
	// log2(searchRange/2)
	EntrySelector uint16
	// (2 * segCount) - searchRange
	RangeShift uint16
	// Ending character code for each segment, last = 0xFFFF.
	EndCode []uint16
	// This value should be zero
	ReservedPad uint16
	// Starting character code for each segment
	StartCode []uint16
	// Delta for all character codes in segment
	IdDelta []uint16
	// Offset in bytes to glyph indexArray, or 0
	IdRangeOffset []uint16

	// format 6
	/**
	UInt16	firstCode	First character code of subrange
	UInt16	entryCount	Number of character codes in subrange
	*/
	FirstCode  uint16
	EntryCount uint16

	// 'cmap' format 8–Mixed 16-bit and 32-bit coverage
	// UInt16	reserved	Set to 0
	Reserved uint16
	// is32:
	// If this value is nonzero, the value of the firstCode field is the lowest character code in subrange,
	Is32    []uint8 // otherwise it is the lowest byte of a two-byte character code range.
	NGroups uint32  // Number of groupings which follow

	// Mixed 16-bit and 32-bit coverage table
	Groups []CmapSubTableFormatMixGroup
}

// required tables
// cmap — character to glyph mapping
type CmapTable struct {
	TTFTable

	Version         uint16 // Version number (Set to zero)
	NumberSubtables uint16 // Number of encoding subtables

	// Encoding subtables
	EncodingSubtables []CmapSubTable
}

/*
*
Fixed	version	0x00010000 if (version 1.0)
Fixed	fontRevision	set by font manufacturer
uint32	checkSumAdjustment	To compute: set it to 0, calculate the checksum for the 'head' table and put it in the table directory, sum the entire font as a uint32_t, then store 0xB1B0AFBA - sum. (The checksum for the 'head' table will be wrong as a result. That is OK; do not reset it.)
uint32	magicNumber	set to 0x5F0F3CF5
uint16	flags	bit 0 - y value of 0 specifies baseline
bit 1 - x position of left most black bit is LSB
bit 2 - scaled point size and actual point size will differ (i.e. 24 point glyph differs from 12 point glyph scaled by factor of 2)
bit 3 - use integer scaling instead of fractional
bit 4 - (used by the Microsoft implementation of the TrueType scaler)
bit 5 - This bit should be set in fonts that are intended to e laid out vertically, and in which the glyphs have been drawn such that an x-coordinate of 0 corresponds to the desired vertical baseline.
bit 6 - This bit must be set to zero.
bit 7 - This bit should be set if the font requires layout for correct linguistic rendering (e.g. Arabic fonts).
bit 8 - This bit should be set for an AAT font which has one or more metamorphosis effects designated as happening by default.
bit 9 - This bit should be set if the font contains any strong right-to-left glyphs.
bit 10 - This bit should be set if the font contains Indic-style rearrangement effects.
bits 11-13 - Defined by Adobe.
bit 14 - This bit should be set if the glyphs in the font are simply generic symbols for code point ranges, such as for a last resort font.
uint16	unitsPerEm	range from 64 to 16384
longDateTime	created	international date
longDateTime	modified	international date
FWord	xMin	for all glyph bounding boxes
FWord	yMin	for all glyph bounding boxes
FWord	xMax	for all glyph bounding boxes
FWord	yMax	for all glyph bounding boxes
uint16	macStyle	bit 0 bold
bit 1 italic
bit 2 underline
bit 3 outline
bit 4 shadow
bit 5 condensed (narrow)
bit 6 extended
uint16	lowestRecPPEM	smallest readable size in pixels
int16	fontDirectionHint	0 Mixed directional glyphs
1 Only strongly left to right glyphs
2 Like 1 but also contains neutrals
-1 Only strongly right to left glyphs
-2 Like -1 but also contains neutrals
int16	indexToLocFormat	0 for short offsets, 1 for long
int16	glyphDataFormat	0 for current format
*/
type HeadTable struct {
	TTFTable
	Version            uint32 // 0x00010000 if (version 1.0)
	FontRevision       uint32 // set by font manufacturer
	CheckSumAdjustment uint32 // To compute: set it to 0, calculate the checksum for the 'head' table and put it in the table directory, sum the entire font as a uint32_t, then store 0xB1B0AFBA - sum. (The checksum for the 'head' table will be wrong as a result. That is OK; do not reset it.)
	MagicNumber        uint32 // set to 0x5F0F3CF5
	Flags              uint16 // bit 0 - y value of 0 specifies baseline
	// bit 1 - x position of left most black bit is LSB
	// bit 2 - scaled point size and actual point size will differ (i.e. 24 point glyph differs from 12 point glyph scaled by factor of 2)
	// bit 3 - use integer scaling instead of fractional
	// bit 4 - (used by the Microsoft implementation of the TrueType scaler)
	// bit 5 - This bit should be set in fonts that are intended to e laid out vertically, and in which the glyphs have been drawn such that an x-coordinate of 0 corresponds to the desired vertical baseline.
	UnitPerEm uint16 // range from 64 to 16384
	Created   uint64 // international date
	Modified  uint64 // international date
	XMin      int16  // for all glyph bounding boxes
	YMin      int16  // for all glyph bounding boxes
	XMax      int16  // for all glyph bounding boxes
	YMax      int16  // for all glyph bounding boxes
	// bit 0 bold
	// bit 1 italic
	// bit 2 underline
	// bit 3 outline
	// bit 4 shadow
	// bit 5 condensed (narrow)
	// bit 6 extended
	MacStyle          uint16 //
	LowestRecPPEM     uint16 // smallest readable size in pixels
	FontDirectionHint int16  // 0 Mixed directional glyphs
	// 1 Only strongly left to right glyphs
	// 2 Like 1 but also contains neutrals
	// -1 Only strongly right to left glyphs
	// -2 Like -1 but also contains neutrals
	IndexToLocFormat int16 // 0 for short offsets, 1 for long
	GlyphDataFormat  int16 // 0 for current format

}

type TTF struct {
	File string

	ScalerType    uint32 // 0x00010000(65546) for TTF or 0x74727565(1953658213) for OTF
	NumTables     uint16 // number of tables
	SearchRange   uint16 // (maximum power of 2 <= numTables)*16
	EntrySelector uint16 // log2(maximum power of 2 <= numTables)
	RangeShift    uint16 // numTables*16-searchRange

	Tables []TTFTableInfo // tables
}

func (ttf *TTF) readTTF() (buf []byte, err error) {
	file, err := os.OpenFile(ttf.File, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}

func (ttf *TTF) readTTFInfo(buf []byte) error {
	ttf.ScalerType = uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	ttf.NumTables = uint16(buf[4])<<8 | uint16(buf[5])
	ttf.SearchRange = uint16(buf[6])<<8 | uint16(buf[7])
	ttf.EntrySelector = uint16(buf[8])<<8 | uint16(buf[9])
	ttf.RangeShift = uint16(buf[10])<<8 | uint16(buf[11])

	if ttf.ScalerType != TTF_MAGIC && ttf.ScalerType != OTF_MAGIC {
		return errors.New("not a ttf or otf file")
	}
	return nil
}

// read cmap table
func readCmapTable(buf []byte, i int) (ti TTFTableInfo) {
	tableInfo := TTFTableInfo{
		Tag:      uint32(buf[12+i*16])<<24 | uint32(buf[13+i*16])<<16 | uint32(buf[14+i*16])<<8 | uint32(buf[15+i*16]),
		CheckSum: uint32(buf[16+i*16])<<24 | uint32(buf[17+i*16])<<16 | uint32(buf[18+i*16])<<8 | uint32(buf[19+i*16]),
		Offset:   uint32(buf[20+i*16])<<24 | uint32(buf[21+i*16])<<16 | uint32(buf[22+i*16])<<8 | uint32(buf[23+i*16]),
		Length:   uint32(buf[24+i*16])<<24 | uint32(buf[25+i*16])<<16 | uint32(buf[26+i*16])<<8 | uint32(buf[27+i*16]),
	}
	// read cmap table
	cmapTable := CmapTable{
		Version:         uint16(buf[tableInfo.Offset+0])<<8 | uint16(buf[tableInfo.Offset+1]),
		NumberSubtables: uint16(buf[tableInfo.Offset+2])<<8 | uint16(buf[tableInfo.Offset+3]),
	}
	// read encoding subtables
	cmapTable.EncodingSubtables = make([]CmapSubTable, cmapTable.NumberSubtables)
	for j := uint32(0); j < uint32(cmapTable.NumberSubtables); j++ {
		encodingSubtable := CmapSubTable{
			PlatformID: uint16(buf[int(tableInfo.Offset+uint32(4)+j*uint32(8))])<<8 | uint16(buf[tableInfo.Offset+5+j*8]),
			EncodingID: uint16(buf[tableInfo.Offset+6+j*8])<<8 | uint16(buf[tableInfo.Offset+7+j*8]),
			SubOffset:  uint32(buf[tableInfo.Offset+8+j*8])<<24 | uint32(buf[tableInfo.Offset+9+j*8])<<16 | uint32(buf[tableInfo.Offset+10+j*8])<<8 | uint32(buf[tableInfo.Offset+11+j*8]),
		}
		// read subtable
		encodingSubtable.Format = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+0])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+1])
		encodingSubtable.Length = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+2])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+3])
		encodingSubtable.Language = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+4])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+5])
		// read glyph index array
		encodingSubtable.GlyphIndexArray = make([]uint8, int(encodingSubtable.Length-6))
		for k := 0; k < int(encodingSubtable.Length-6); k++ {
			index := int(tableInfo.Offset) + int(encodingSubtable.SubOffset) + 6 + k
			encodingSubtable.GlyphIndexArray[k] = buf[index]
		}
		// read subtable
		switch encodingSubtable.Format {
		case 2:
			encodingSubtable.SubHeaderKeys = make([]uint16, uint32(encodingSubtable.Length)-6)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.Length)-6); k++ {
				encodingSubtable.SubHeaderKeys[k] = uint16(buf[int(tableInfo.Offset+encodingSubtable.SubOffset+6+k)])
			}
			// read subHeaders
			encodingSubtable.SubHeaders = make([]CmapSubHeader, uint32(encodingSubtable.Length)-6)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.Length)-6); k++ {
				encodingSubtable.SubHeaders[k].FirstCode = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6+k])
				encodingSubtable.SubHeaders[k].EntryCount = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6+k])
				encodingSubtable.SubHeaders[k].IdDelta = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6+k])
				encodingSubtable.SubHeaders[k].IdRangeOffset = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6+k])
			}
			// read glyph index array
			encodingSubtable.GlyphIndexArray16 = make([]uint16, uint32(encodingSubtable.Length)-6)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.Length)-6); k++ {
				encodingSubtable.GlyphIndexArray16[k] = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6+k])
			}
		case 4:
			encodingSubtable.SegCountX2 = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+7])
			encodingSubtable.SearchRange = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+8])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+9])
			encodingSubtable.EntrySelector = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+10])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+11])
			encodingSubtable.RangeShift = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+12])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+13])
			// read end code
			encodingSubtable.EndCode = make([]uint16, uint32(encodingSubtable.SegCountX2)/2)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.SegCountX2)/2); k++ {
				encodingSubtable.EndCode[k] = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+14+k*2])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+15+k*2])
			}
			encodingSubtable.ReservedPad = uint16(buf[int(tableInfo.Offset+encodingSubtable.SubOffset+14+uint32(uint32(encodingSubtable.SegCountX2)/2*2))])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+15+uint32(encodingSubtable.SegCountX2)/2*2])
			// read start code
			encodingSubtable.StartCode = make([]uint16, uint32(encodingSubtable.SegCountX2)/2)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.SegCountX2)/2); k++ {
				encodingSubtable.StartCode[k] = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+16+uint32(uint32(encodingSubtable.SegCountX2)/2*2)+k*2])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+17+uint32(encodingSubtable.SegCountX2)/2*2+k*2])
			}
			// read id delta
			encodingSubtable.IdDelta = make([]uint16, uint32(encodingSubtable.SegCountX2)/2)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.SegCountX2)/2); k++ {
				encodingSubtable.IdDelta[k] = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+16+uint32(uint32(encodingSubtable.SegCountX2)/2*4)+k*2])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+17+uint32(encodingSubtable.SegCountX2)/2*4+k*2])
			}
			// read id range offset
			encodingSubtable.IdRangeOffset = make([]uint16, uint32(encodingSubtable.SegCountX2)/2)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.SegCountX2)/2); k++ {
				encodingSubtable.IdRangeOffset[k] = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+16+uint32(uint32(encodingSubtable.SegCountX2)/2*6)+k*2])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+17+uint32(encodingSubtable.SegCountX2)/2*6+k*2])
			}
		case 6:
			encodingSubtable.FirstCode = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+7])
			encodingSubtable.EntryCount = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+8])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+9])
		case 8:
			encodingSubtable.Reserved = uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+6])<<8 | uint16(buf[tableInfo.Offset+encodingSubtable.SubOffset+7])
			encodingSubtable.Is32 = make([]uint8, uint32(encodingSubtable.Length)-6)
			for k := uint32(0); k < uint32(uint32(encodingSubtable.Length)-6); k++ {
				encodingSubtable.Is32[k] = buf[tableInfo.Offset+encodingSubtable.SubOffset+8+k]
			}
			encodingSubtable.NGroups = uint32(buf[int(tableInfo.Offset+encodingSubtable.SubOffset+8+uint32(uint32(encodingSubtable.Length))-6)])<<24 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+9+uint32(encodingSubtable.Length)-6])<<16 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+10+uint32(encodingSubtable.Length)-6])<<8 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+11+uint32(encodingSubtable.Length)-6])
			// read groups
			encodingSubtable.Groups = make([]CmapSubTableFormatMixGroup, encodingSubtable.NGroups)
			for k := uint32(0); k < uint32(encodingSubtable.NGroups); k++ {
				encodingSubtable.Groups[k].StartCharCode = uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+12+uint32(encodingSubtable.Length)-6+k*12])<<24 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+13+uint32(encodingSubtable.Length)-6+k*12])<<16 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+14+uint32(encodingSubtable.Length)-6+k*12])<<8 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+15+uint32(encodingSubtable.Length)-6+k*12])
				encodingSubtable.Groups[k].EndCharCode = uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+16+uint32(encodingSubtable.Length)-6+k*12])<<24 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+17+uint32(encodingSubtable.Length)-6+k*12])<<16 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+18+uint32(encodingSubtable.Length)-6+k*12])<<8 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+19+uint32(encodingSubtable.Length)-6+k*12])
				encodingSubtable.Groups[k].StartGlyphCode = uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+20+uint32(encodingSubtable.Length)-6+k*12])<<24 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+21+uint32(encodingSubtable.Length)-6+k*12])<<16 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+22+uint32(encodingSubtable.Length)-6+k*12])<<8 | uint32(buf[tableInfo.Offset+encodingSubtable.SubOffset+23+uint32(encodingSubtable.Length)-6+k*12])
			}
		}
		// append encoding subtable
		cmapTable.EncodingSubtables[j] = encodingSubtable
	}
	tableInfo.Table = cmapTable
	return tableInfo
}

// read head table
func readHeadTable(buf []byte, i int) (ti TTFTableInfo) {
	ti = TTFTableInfo{
		Tag:      uint32(buf[12+i*16])<<24 | uint32(buf[13+i*16])<<16 | uint32(buf[14+i*16])<<8 | uint32(buf[15+i*16]),
		CheckSum: uint32(buf[16+i*16])<<24 | uint32(buf[17+i*16])<<16 | uint32(buf[18+i*16])<<8 | uint32(buf[19+i*16]),
		Offset:   uint32(buf[20+i*16])<<24 | uint32(buf[21+i*16])<<16 | uint32(buf[22+i*16])<<8 | uint32(buf[23+i*16]),
		Length:   uint32(buf[24+i*16])<<24 | uint32(buf[25+i*16])<<16 | uint32(buf[26+i*16])<<8 | uint32(buf[27+i*16]),

		Table: HeadTable{
			Version:            uint32(buf[ti.Offset+0])<<24 | uint32(buf[ti.Offset+1])<<16 | uint32(buf[ti.Offset+2])<<8 | uint32(buf[ti.Offset+3]),
			FontRevision:       uint32(buf[ti.Offset+4])<<24 | uint32(buf[ti.Offset+5])<<16 | uint32(buf[ti.Offset+6])<<8 | uint32(buf[ti.Offset+7]),
			CheckSumAdjustment: uint32(buf[ti.Offset+8])<<24 | uint32(buf[ti.Offset+9])<<16 | uint32(buf[ti.Offset+10])<<8 | uint32(buf[ti.Offset+11]),
			MagicNumber:        uint32(buf[ti.Offset+12])<<24 | uint32(buf[ti.Offset+13])<<16 | uint32(buf[ti.Offset+14])<<8 | uint32(buf[ti.Offset+15]),
			Flags:              uint16(buf[ti.Offset+16])<<8 | uint16(buf[ti.Offset+17]),
			UnitPerEm:          uint16(buf[ti.Offset+18])<<8 | uint16(buf[ti.Offset+19]),
			Created:            uint64(buf[ti.Offset+20])<<56 | uint64(buf[ti.Offset+21])<<48 | uint64(buf[ti.Offset+22])<<40 | uint64(buf[ti.Offset+23])<<32 | uint64(buf[ti.Offset+24])<<24 | uint64(buf[ti.Offset+25])<<16 | uint64(buf[ti.Offset+26])<<8 | uint64(buf[ti.Offset+27]),
			Modified:           uint64(buf[ti.Offset+28])<<56 | uint64(buf[ti.Offset+29])<<48 | uint64(buf[ti.Offset+30])<<40 | uint64(buf[ti.Offset+31])<<32 | uint64(buf[ti.Offset+32])<<24 | uint64(buf[ti.Offset+33])<<16 | uint64(buf[ti.Offset+34])<<8 | uint64(buf[ti.Offset+35]),
			XMin:               int16(buf[ti.Offset+36])<<8 | int16(buf[ti.Offset+37]),
			YMin:               int16(buf[ti.Offset+38])<<8 | int16(buf[ti.Offset+39]),
			XMax:               int16(buf[ti.Offset+40])<<8 | int16(buf[ti.Offset+41]),
			YMax:               int16(buf[ti.Offset+42])<<8 | int16(buf[ti.Offset+43]),
			MacStyle:           uint16(buf[ti.Offset+44])<<8 | uint16(buf[ti.Offset+45]),
			LowestRecPPEM:      uint16(buf[ti.Offset+46])<<8 | uint16(buf[ti.Offset+47]),
			FontDirectionHint:  int16(buf[ti.Offset+48])<<8 | int16(buf[ti.Offset+49]),
			IndexToLocFormat:   int16(buf[ti.Offset+50])<<8 | int16(buf[ti.Offset+51]),
			GlyphDataFormat:    int16(buf[ti.Offset+52])<<8 | int16(buf[ti.Offset+53]),
		},
	}
	return ti
}

func (ttf *TTF) readTTFTables(buf []byte) error {
	for i := 0; i < int(ttf.NumTables); i++ {
		tag := string([]byte{buf[12+i*16], buf[13+i*16], buf[14+i*16], buf[15+i*16]})
		switch tag {
		case "cmap":
			tableInfo := readCmapTable(buf, i)
			ttf.Tables = append(ttf.Tables, tableInfo)
			break
		case "head":
			tableInfo := readHeadTable(buf, i)
			ttf.Tables = append(ttf.Tables, tableInfo)
			break
		}
	}
	return nil
}

func NewTTF(fileName string) (*TTF, error) {
	ttf := &TTF{
		File:   fileName,
		Tables: make([]TTFTableInfo, 0),
	}
	buf, err := ttf.readTTF()
	if err != nil {
		return nil, err
	}
	// header
	err = ttf.readTTFInfo(buf)
	if err != nil {
		return nil, err
	}
	// tables
	err = ttf.readTTFTables(buf)
	if err != nil {
		return nil, err
	}
	return ttf, nil
}

func PrintTagName(tag uint32) string {
	return string([]byte{byte(tag >> 24), byte(tag >> 16), byte(tag >> 8), byte(tag)})
}
