package objfile

import (
	"encoding/binary"
	"reflect"
	"testing"
)

// createTestModuleData creates a byte slice containing test moduledata for a specific version and architecture
func createTestModuleData(version string, is64bit bool) []byte {
	// Create a buffer large enough to hold the largest moduledata structure
	buf := make([]byte, 1024)
	
	// Get the layout for this version
	layout := getLayout(version, is64bit)
	
	// Helper function to write a value at an offset
	writeAt := func(offset FieldOffset, value uint64) {
		if offset.Size == 0 {
			return // Skip writing if field doesn't exist
		}
		if is64bit {
			binary.LittleEndian.PutUint64(buf[offset.Offset:], value)
		} else {
			binary.LittleEndian.PutUint32(buf[offset.Offset:], uint32(value))
		}
	}
	
	// Helper function to write a slice at an offset
	writeSlice := func(offset FieldOffset, data, length, capacity uint64) {
		if offset.Size == 0 {
			return // Skip writing if field doesn't exist
		}
		if is64bit {
			binary.LittleEndian.PutUint64(buf[offset.Offset:], data)
			binary.LittleEndian.PutUint64(buf[offset.Offset+8:], length)
			binary.LittleEndian.PutUint64(buf[offset.Offset+16:], capacity)
		} else {
			binary.LittleEndian.PutUint32(buf[offset.Offset:], uint32(data))
			binary.LittleEndian.PutUint32(buf[offset.Offset+4:], uint32(length))
			binary.LittleEndian.PutUint32(buf[offset.Offset+8:], uint32(capacity))
		}
	}
	
	// Write common fields
	writeAt(layout.Text, 0x1000)
	writeAt(layout.Types, 0x2000)
	writeAt(layout.ETypes, 0x3000)
	writeAt(layout.Minpc, 0x4000)
	
	// Write slice fields
	writeSlice(layout.Typelinks, 0x5000, 10, 20)
	writeSlice(layout.ITablinks, 0x6000, 15, 25)
	writeSlice(layout.Ftab, 0x7000, 20, 30)
	
	// Write version-specific fields
	if version == "1.5" {
		writeSlice(layout.LegacyTypes, 0x8000, 10, 20)
	} else if version >= "1.18" {
		writeAt(layout.Rodata, 0x9000)
		writeAt(layout.Gofunc, 0xa000)
		if version >= "1.20" {
			writeAt(layout.Covctrs, 0xb000)
			writeAt(layout.Ecovctrs, 0xc000)
			if version >= "1.22" {
				writeSlice(layout.InitTasks, 0xd000, 5, 10)
			}
		}
	}
	
	return buf
}

func TestModuleDataParsing(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		is64bit  bool
		expected *ModuleData
	}{
		{
			name:    "Go 1.5 32-bit",
			version: "1.5",
			is64bit: false,
			expected: &ModuleData{
				TextVA:      0x1000,
				Types:       0x2000,
				ETypes:      0x3000,
				Typelinks:   GoSlice64{Data: 0x5000, Len: 10, Capacity: 20},
				ITablinks:   GoSlice64{Data: 0x6000, Len: 15, Capacity: 25},
				LegacyTypes: GoSlice64{Data: 0x8000, Len: 10, Capacity: 20},
			},
		},
		{
			name:    "Go 1.5 64-bit",
			version: "1.5",
			is64bit: true,
			expected: &ModuleData{
				TextVA:      0x1000,
				Types:       0x2000,
				ETypes:      0x3000,
				Typelinks:   GoSlice64{Data: 0x5000, Len: 10, Capacity: 20},
				ITablinks:   GoSlice64{Data: 0x6000, Len: 15, Capacity: 25},
				LegacyTypes: GoSlice64{Data: 0x8000, Len: 10, Capacity: 20},
			},
		},
		{
			name:    "Go 1.18 32-bit",
			version: "1.18",
			is64bit: false,
			expected: &ModuleData{
				TextVA:    0x1000,
				Types:     0x2000,
				ETypes:    0x3000,
				Typelinks: GoSlice64{Data: 0x5000, Len: 10, Capacity: 20},
				ITablinks: GoSlice64{Data: 0x6000, Len: 15, Capacity: 25},
				Rodata:    0x9000,
				Gofunc:    0xa000,
			},
		},
		{
			name:    "Go 1.18 64-bit",
			version: "1.18",
			is64bit: true,
			expected: &ModuleData{
				TextVA:    0x1000,
				Types:     0x2000,
				ETypes:    0x3000,
				Typelinks: GoSlice64{Data: 0x5000, Len: 10, Capacity: 20},
				ITablinks: GoSlice64{Data: 0x6000, Len: 15, Capacity: 25},
				Rodata:    0x9000,
				Gofunc:    0xa000,
			},
		},
		{
			name:    "Go 1.22 64-bit",
			version: "1.22",
			is64bit: true,
			expected: &ModuleData{
				TextVA:     0x1000,
				Types:      0x2000,
				ETypes:     0x3000,
				Typelinks:  GoSlice64{Data: 0x5000, Len: 10, Capacity: 20},
				ITablinks:  GoSlice64{Data: 0x6000, Len: 15, Capacity: 25},
				Rodata:     0x9000,
				Gofunc:     0xa000,
				Covctrs:    0xb000,
				Ecovctrs:   0xc000,
				InitTasks:  GoSlice64{Data: 0xd000, Len: 5, Capacity: 10},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test data
			data := createTestModuleData(tt.version, tt.is64bit)
			
			// Parse the data
			md, err := ParseModuleData(data, tt.version, tt.is64bit, true)
			if err != nil {
				t.Fatalf("ParseModuleData failed: %v", err)
			}
			
			// Compare with expected values
			if !reflect.DeepEqual(md, tt.expected) {
				t.Errorf("got %+v, want %+v", md, tt.expected)
			}
		})
	}
}

func TestModuleDataValidation(t *testing.T) {
	// Create test data for Go 1.22 64-bit
	data := createTestModuleData("1.22", true)
	
	// Parse the data
	md, err := ParseModuleData(data, "1.22", true, true)
	if err != nil {
		t.Fatalf("ParseModuleData failed: %v", err)
	}
	
	// Test validation with matching first function entry
	err = ValidateModuleData(md, 0x1000) // matches TextVA
	if err != nil {
		t.Errorf("Validation failed with matching first function entry: %v", err)
	}
	
	// Test validation with non-matching first function entry
	err = ValidateModuleData(md, 0x5000) // doesn't match TextVA
	if err == nil {
		t.Error("Validation succeeded with non-matching first function entry")
	}
} 