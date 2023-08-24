package fontcompress_test

import (
	"testing"

	font_compress "github.com/RustynailPlease/fontcompress"
)

func TestReadTTFInfo(t *testing.T) {
	ttf, err := font_compress.NewTTF("./fonts/LXGWWenKai-Regular.ttf")
	if err != nil {
		t.Error(err.Error())
	}
	// t.Log("ttf tables:", ttf.Tables)
	t.Log("ttf scaler type:", ttf.ScalerType)
	t.Log("ttf num tables:", ttf.NumTables)
	t.Log("ttf search range:", ttf.SearchRange)
	t.Log("ttf entry selector:", ttf.EntrySelector)
	t.Log("ttf range shift:", ttf.RangeShift)
	// table tags
	for _, table := range ttf.Tables {
		t.Log("=====================================================")
		t.Log("table tag:", font_compress.PrintTagName(table.Tag))
		t.Log("table check sum:", table.CheckSum)
		t.Log("table offset:", table.Offset)
		t.Log("table length:", table.Length)
		switch font_compress.PrintTagName(table.Tag) {
		case "cmap":
			cmap := table.Table.(font_compress.CmapTable)
			t.Log("cmap version:", cmap.Version)
			for _, sub := range cmap.EncodingSubtables {
				t.Log("cmap subtable platform ID:", sub.PlatformID)
				t.Log("cmap subtable encoding ID:", sub.EncodingID)
				t.Log("cmap subtable Language:", sub.Language)
				t.Log("cmap subtable format:", sub.Format)
			}
			break
		case "head":
			head := table.Table.(font_compress.HeadTable)
			t.Log("head version:", head.Version)
			t.Log("head font revision:", head.FontRevision)
			t.Log("head check sum adjustment:", head.CheckSumAdjustment)
			t.Log("head magic number:", head.MagicNumber)
			t.Log("head flags:", head.Flags)
			t.Log("head units perm:", head.UnitPerEm)
			t.Log("head created:", head.Created)
			t.Log("head modified:", head.Modified)
			t.Log("head x min:", head.XMin)
			t.Log("head y min:", head.YMin)
			t.Log("head x max:", head.XMax)
			t.Log("head y max:", head.YMax)
			t.Log("head mac style:", head.MacStyle)
			t.Log("head lowest rec PPEM:", head.LowestRecPPEM)
			t.Log("head font direction hint:", head.FontDirectionHint)
			t.Log("head index to loc format:", head.IndexToLocFormat)
			t.Log("head glyph data format:", head.GlyphDataFormat)
			break
		default:
			break
		}
	}
}
