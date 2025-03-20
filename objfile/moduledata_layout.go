package objfile

import (
	"encoding/binary"
	"fmt"
)

// FieldOffset represents the offset of a field in the moduledata structure
type FieldOffset struct {
	Offset uint64
	Size   uint64
}

// ModuleDataLayout defines the layout of fields in the moduledata structure
type ModuleDataLayout struct {
	// Common fields across versions
	Text       FieldOffset
	Types      FieldOffset
	ETypes     FieldOffset
	Typelinks  FieldOffset
	ITablinks  FieldOffset
	Ftab       FieldOffset
	Minpc      FieldOffset
	Textsectmap FieldOffset

	// Version-specific fields
	LegacyTypes FieldOffset // Only used in Go 1.5-1.6
	Rodata      FieldOffset // Added in Go 1.18
	Gofunc      FieldOffset // Added in Go 1.18
	Covctrs     FieldOffset // Added in Go 1.20
	Ecovctrs    FieldOffset // Added in Go 1.20
	InitTasks   FieldOffset // Added in Go 1.21

	// Pointer size for this layout
	PtrSize uint64
}

// versionLayoutMap maps Go versions to their corresponding moduledata layouts
var versionLayoutMap = map[string]ModuleDataLayout{
	// Go 1.5-1.6 layout (32-bit)
	"1.5": {
		Text:       FieldOffset{Offset: 0x40, Size: 4},
		Types:      FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		ETypes:     FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		Typelinks:  FieldOffset{Offset: 0x60, Size: 12},
		ITablinks:  FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		Ftab:       FieldOffset{Offset: 0x8, Size: 12},
		Minpc:      FieldOffset{Offset: 0x20, Size: 4},
		Textsectmap: FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		LegacyTypes: FieldOffset{Offset: 0x60, Size: 12},
		PtrSize:    4,
	},

	// Go 1.5-1.6 layout (64-bit)
	"1.5_64": {
		Text:       FieldOffset{Offset: 0x80, Size: 8},
		Types:      FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		ETypes:     FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		Typelinks:  FieldOffset{Offset: 0xc0, Size: 24},
		ITablinks:  FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		Ftab:       FieldOffset{Offset: 0x10, Size: 24},
		Minpc:      FieldOffset{Offset: 0x40, Size: 8},
		Textsectmap: FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.5
		LegacyTypes: FieldOffset{Offset: 0xc0, Size: 24},
		PtrSize:    8,
	},

	// Go 1.7-1.17 layout (32-bit)
	"1.7": {
		Text:       FieldOffset{Offset: 0x40, Size: 4},
		Types:      FieldOffset{Offset: 0x60, Size: 4},
		ETypes:     FieldOffset{Offset: 0x64, Size: 4},
		Typelinks:  FieldOffset{Offset: 0x68, Size: 12},
		ITablinks:  FieldOffset{Offset: 0x74, Size: 12},
		Ftab:       FieldOffset{Offset: 0x8, Size: 12},
		Minpc:      FieldOffset{Offset: 0x20, Size: 4},
		Textsectmap: FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.7
		PtrSize:    4,
	},

	// Go 1.7-1.17 layout (64-bit)
	"1.7_64": {
		Text:       FieldOffset{Offset: 0x80, Size: 8},
		Types:      FieldOffset{Offset: 0xc0, Size: 8},
		ETypes:     FieldOffset{Offset: 0xc8, Size: 8},
		Typelinks:  FieldOffset{Offset: 0xd0, Size: 24},
		ITablinks:  FieldOffset{Offset: 0xe8, Size: 24},
		Ftab:       FieldOffset{Offset: 0x10, Size: 24},
		Minpc:      FieldOffset{Offset: 0x40, Size: 8},
		Textsectmap: FieldOffset{Offset: 0x0, Size: 0}, // Not present in 1.7
		PtrSize:    8,
	},

	// Go 1.18-1.19 layout (32-bit)
	"1.18": {
		Text:       FieldOffset{Offset: 0x40, Size: 4},
		Types:      FieldOffset{Offset: 0x60, Size: 4},
		ETypes:     FieldOffset{Offset: 0x64, Size: 4},
		Typelinks:  FieldOffset{Offset: 0x68, Size: 12},
		ITablinks:  FieldOffset{Offset: 0x74, Size: 12},
		Ftab:       FieldOffset{Offset: 0x8, Size: 12},
		Minpc:      FieldOffset{Offset: 0x20, Size: 4},
		Textsectmap: FieldOffset{Offset: 0x80, Size: 12},
		Rodata:     FieldOffset{Offset: 0x70, Size: 4},
		Gofunc:     FieldOffset{Offset: 0x74, Size: 4},
		PtrSize:    4,
	},

	// Go 1.18-1.19 layout (64-bit)
	"1.18_64": {
		Text:       FieldOffset{Offset: 0x80, Size: 8},
		Types:      FieldOffset{Offset: 0xc0, Size: 8},
		ETypes:     FieldOffset{Offset: 0xc8, Size: 8},
		Typelinks:  FieldOffset{Offset: 0xd0, Size: 24},
		ITablinks:  FieldOffset{Offset: 0xe8, Size: 24},
		Ftab:       FieldOffset{Offset: 0x10, Size: 24},
		Minpc:      FieldOffset{Offset: 0x40, Size: 8},
		Textsectmap: FieldOffset{Offset: 0x100, Size: 24},
		Rodata:     FieldOffset{Offset: 0xe0, Size: 8},
		Gofunc:     FieldOffset{Offset: 0xe8, Size: 8},
		PtrSize:    8,
	},

	// Go 1.20-1.21 layout (32-bit)
	"1.20": {
		Text:       FieldOffset{Offset: 0x40, Size: 4},
		Types:      FieldOffset{Offset: 0x60, Size: 4},
		ETypes:     FieldOffset{Offset: 0x64, Size: 4},
		Typelinks:  FieldOffset{Offset: 0x68, Size: 12},
		ITablinks:  FieldOffset{Offset: 0x74, Size: 12},
		Ftab:       FieldOffset{Offset: 0x8, Size: 12},
		Minpc:      FieldOffset{Offset: 0x20, Size: 4},
		Textsectmap: FieldOffset{Offset: 0x80, Size: 12},
		Rodata:     FieldOffset{Offset: 0x70, Size: 4},
		Gofunc:     FieldOffset{Offset: 0x74, Size: 4},
		Covctrs:    FieldOffset{Offset: 0x78, Size: 4},
		Ecovctrs:   FieldOffset{Offset: 0x7c, Size: 4},
		PtrSize:    4,
	},

	// Go 1.20-1.21 layout (64-bit)
	"1.20_64": {
		Text:       FieldOffset{Offset: 0x80, Size: 8},
		Types:      FieldOffset{Offset: 0xc0, Size: 8},
		ETypes:     FieldOffset{Offset: 0xc8, Size: 8},
		Typelinks:  FieldOffset{Offset: 0xd0, Size: 24},
		ITablinks:  FieldOffset{Offset: 0xe8, Size: 24},
		Ftab:       FieldOffset{Offset: 0x10, Size: 24},
		Minpc:      FieldOffset{Offset: 0x40, Size: 8},
		Textsectmap: FieldOffset{Offset: 0x100, Size: 24},
		Rodata:     FieldOffset{Offset: 0xe0, Size: 8},
		Gofunc:     FieldOffset{Offset: 0xe8, Size: 8},
		Covctrs:    FieldOffset{Offset: 0xf0, Size: 8},
		Ecovctrs:   FieldOffset{Offset: 0xf8, Size: 8},
		PtrSize:    8,
	},

	// Go 1.22 layout (32-bit)
	"1.22": {
		Text:       FieldOffset{Offset: 0x40, Size: 4},
		Types:      FieldOffset{Offset: 0x60, Size: 4},
		ETypes:     FieldOffset{Offset: 0x64, Size: 4},
		Typelinks:  FieldOffset{Offset: 0x68, Size: 12},
		ITablinks:  FieldOffset{Offset: 0x74, Size: 12},
		Ftab:       FieldOffset{Offset: 0x8, Size: 12},
		Minpc:      FieldOffset{Offset: 0x20, Size: 4},
		Textsectmap: FieldOffset{Offset: 0x80, Size: 12},
		Rodata:     FieldOffset{Offset: 0x70, Size: 4},
		Gofunc:     FieldOffset{Offset: 0x74, Size: 4},
		Covctrs:    FieldOffset{Offset: 0x78, Size: 4},
		Ecovctrs:   FieldOffset{Offset: 0x7c, Size: 4},
		InitTasks:  FieldOffset{Offset: 0x8c, Size: 12},
		PtrSize:    4,
	},

	// Go 1.22 layout (64-bit)
	"1.22_64": {
		Text:       FieldOffset{Offset: 0x80, Size: 8},
		Types:      FieldOffset{Offset: 0xc0, Size: 8},
		ETypes:     FieldOffset{Offset: 0xc8, Size: 8},
		Typelinks:  FieldOffset{Offset: 0xd0, Size: 24},
		ITablinks:  FieldOffset{Offset: 0xe8, Size: 24},
		Ftab:       FieldOffset{Offset: 0x10, Size: 24},
		Minpc:      FieldOffset{Offset: 0x40, Size: 8},
		Textsectmap: FieldOffset{Offset: 0x100, Size: 24},
		Rodata:     FieldOffset{Offset: 0xe0, Size: 8},
		Gofunc:     FieldOffset{Offset: 0xe8, Size: 8},
		Covctrs:    FieldOffset{Offset: 0xf0, Size: 8},
		Ecovctrs:   FieldOffset{Offset: 0xf8, Size: 8},
		InitTasks:  FieldOffset{Offset: 0x118, Size: 24},
		PtrSize:    8,
	},
}

// getLayout returns the appropriate ModuleDataLayout for the given version and architecture
func getLayout(version string, is64bit bool) ModuleDataLayout {
	key := version
	if is64bit {
		key += "_64"
	}
	if layout, ok := versionLayoutMap[key]; ok {
		return layout
	}
	// Return a default layout if version not found
	return ModuleDataLayout{
		PtrSize: 4,
	}
}

// readField reads a field from the moduledata byte slice using the given offset
func readField(data []byte, offset FieldOffset, littleEndian bool) (uint64, error) {
	if offset.Size == 0 {
		return 0, nil
	}
	if offset.Offset+offset.Size > uint64(len(data)) {
		return 0, fmt.Errorf("field offset %d + size %d exceeds data length %d", offset.Offset, offset.Size, len(data))
	}
	fieldData := data[offset.Offset : offset.Offset+offset.Size]
	if offset.Size == 8 {
		if littleEndian {
			return binary.LittleEndian.Uint64(fieldData), nil
		}
		return binary.BigEndian.Uint64(fieldData), nil
	} else if offset.Size == 4 {
		if littleEndian {
			return uint64(binary.LittleEndian.Uint32(fieldData)), nil
		}
		return uint64(binary.BigEndian.Uint32(fieldData)), nil
	}
	return 0, fmt.Errorf("unsupported field size: %d", offset.Size)
}

// readSlice reads a GoSlice from the moduledata byte slice using the given offset
func readSlice(data []byte, offset FieldOffset, ptrSize uint64, littleEndian bool) (GoSlice64, error) {
	if offset.Size == 0 {
		return GoSlice64{}, nil
	}
	if offset.Offset+offset.Size > uint64(len(data)) {
		return GoSlice64{}, fmt.Errorf("slice offset %d + size %d exceeds data length %d", offset.Offset, offset.Size, len(data))
	}
	sliceData := data[offset.Offset : offset.Offset+offset.Size]
	
	var slice GoSlice64
	if ptrSize == 8 {
		if littleEndian {
			slice.Data = pvoid64(binary.LittleEndian.Uint64(sliceData[0:8]))
			slice.Len = binary.LittleEndian.Uint64(sliceData[8:16])
			slice.Capacity = binary.LittleEndian.Uint64(sliceData[16:24])
		} else {
			slice.Data = pvoid64(binary.BigEndian.Uint64(sliceData[0:8]))
			slice.Len = binary.BigEndian.Uint64(sliceData[8:16])
			slice.Capacity = binary.BigEndian.Uint64(sliceData[16:24])
		}
	} else {
		if littleEndian {
			slice.Data = pvoid64(binary.LittleEndian.Uint32(sliceData[0:4]))
			slice.Len = uint64(binary.LittleEndian.Uint32(sliceData[4:8]))
			slice.Capacity = uint64(binary.LittleEndian.Uint32(sliceData[8:12]))
		} else {
			slice.Data = pvoid64(binary.BigEndian.Uint32(sliceData[0:4]))
			slice.Len = uint64(binary.BigEndian.Uint32(sliceData[4:8]))
			slice.Capacity = uint64(binary.BigEndian.Uint32(sliceData[8:12]))
		}
	}
	return slice, nil
} 