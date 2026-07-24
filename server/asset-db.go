package remote_ui_server

import (
	"bytes"

	"github.com/jurgen-kluft/go-datastream/codestream"
	fontpack "github.com/jurgen-kluft/go-gx2/fontpak"
	spritepack "github.com/jurgen-kluft/go-gx2/spritepak"
)

type PalettePack struct {
	Palettes []spritepack.PaletteRGB565
}

type AssetDb struct {
	FontPack    []fontpack.Font
	SpritePack  []spritepack.Sprite
	PalettePack [][]uint16
}

func BuildAssetDatabase(spritePackCfgFilepath string, fontPackCfgFilepath string) (*AssetDb, error) {

	// Load all sprite packs and font packs into memory, so they can be shared across user interfaces.
	spritePackCfg, err := spritepack.LoadConfig(spritePackCfgFilepath)
	if err != nil {
		return nil, err
	}
	sprites, rgbPalettes, err := spritepack.Build(spritePackCfg)
	if err != nil {
		return nil, err
	}

	palettes := make([][]uint16, len(rgbPalettes))
	for i, palette := range rgbPalettes {
		palettes[i] = make([]uint16, len(palette))
		for j, color := range palette {
			palettes[i][j] = color.ToRGB565()
		}
	}

	fontPackCfg, err := fontpack.LoadConfig(fontPackCfgFilepath)
	if err != nil {
		return nil, err
	}

	fontPack, err := fontpack.Build(fontPackCfg)
	if err != nil {
		return nil, err
	}

	// Build the AssetDb
	assetDatabase := &AssetDb{
		SpritePack:  sprites,
		PalettePack: palettes,
		FontPack:    fontPack.Fonts,
	}

	return assetDatabase, nil
}

func (db *AssetDb) GetBytes() []byte {

	bufWriter := bytes.NewBuffer(nil)
	codestream.WriteToStream(bufWriter, *db)

	return bufWriter.Bytes()
}
