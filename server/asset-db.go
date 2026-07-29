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

type AssetEntry struct {
	Name        string
	FontPack    []fontpack.Font
	SpritePack  []spritepack.Sprite
	PalettePack [][]uint16
	Script      []byte
}

type AssetDb struct {
	Entries  []AssetEntry
	AssetDbs map[string][]byte
}

func BuildAssetDatabase(assets map[string]AssetDescriptor) (*AssetDb, error) {
	db := &AssetDb{Entries: []AssetEntry{}}
	for _, a := range assets {
		spritePackCfgFilepath := a.SpritePackCfgFile
		fontPackCfgFilepath := a.FontPackCfgFile

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

		// Compile the script file into a byte slice
		// TODO

		// Build the AssetDb
		asset := AssetEntry{
			Name:        a.Name,
			SpritePack:  sprites,
			PalettePack: palettes,
			FontPack:    fontPack.Fonts,
			Script:      []byte{},
		}

		db.Entries = append(db.Entries, asset)
	}

	for _, a := range db.Entries {
		bufWriter := bytes.NewBuffer(nil)
		codestream.WriteToStream(bufWriter, a)
		db.AssetDbs[a.Name] = bufWriter.Bytes()
	}

	return db, nil
}

func (db *AssetDb) GetBytesFor(name string) []byte {
	if assets, ok := db.AssetDbs[name]; ok {
		return assets
	}
	return nil
}

func (db *AssetDb) GetScriptBytesFor(name string) []byte {
	return nil
}
