package objfile

import (
	"fmt"
	"strings"
)

// ParseModuleData parses the moduledata structure using the offset-based approach
func ParseModuleData(data []byte, version string, is64bit bool, littleEndian bool) (*ModuleData, error) {
	layout := getLayout(version, is64bit)
	
	md := &ModuleData{}
	
	// Parse common fields
	var err error
	md.TextVA, err = readField(data, layout.Text, littleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to read TextVA field: %v", err)
	}
	
	md.Types, err = readField(data, layout.Types, littleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to read Types field: %v", err)
	}
	
	md.ETypes, err = readField(data, layout.ETypes, littleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to read ETypes field: %v", err)
	}
	
	md.Typelinks, err = readSlice(data, layout.Typelinks, layout.PtrSize, littleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to read Typelinks field: %v", err)
	}
	
	md.ITablinks, err = readSlice(data, layout.ITablinks, layout.PtrSize, littleEndian)
	if err != nil {
		return nil, fmt.Errorf("failed to read ITablinks field: %v", err)
	}
	
	// Parse version-specific fields
	if strings.HasPrefix(version, "1.5") {
		// Go 1.5-1.6 specific fields
		md.LegacyTypes, err = readSlice(data, layout.LegacyTypes, layout.PtrSize, littleEndian)
		if err != nil {
			return nil, fmt.Errorf("failed to read LegacyTypes field: %v", err)
		}
	} else if strings.Compare(version, "1.7") >= 0 {
		// Go 1.7+ specific fields
		md.Rodata, err = readField(data, layout.Rodata, littleEndian)
		if err != nil {
			return nil, fmt.Errorf("failed to read Rodata field: %v", err)
		}
		
		md.Gofunc, err = readField(data, layout.Gofunc, littleEndian)
		if err != nil {
			return nil, fmt.Errorf("failed to read Gofunc field: %v", err)
		}
		
		if strings.Compare(version, "1.20") >= 0 {
			// Go 1.20+ specific fields
			md.Covctrs, err = readField(data, layout.Covctrs, littleEndian)
			if err != nil {
				return nil, fmt.Errorf("failed to read Covctrs field: %v", err)
			}
			
			md.Ecovctrs, err = readField(data, layout.Ecovctrs, littleEndian)
			if err != nil {
				return nil, fmt.Errorf("failed to read Ecovctrs field: %v", err)
			}
			
			if strings.Compare(version, "1.22") >= 0 {
				// Go 1.22+ specific fields
				md.InitTasks, err = readSlice(data, layout.InitTasks, layout.PtrSize, littleEndian)
				if err != nil {
					return nil, fmt.Errorf("failed to read InitTasks field: %v", err)
				}
			}
		}
	}
	
	return md, nil
}

// ValidateModuleData performs validation checks on the parsed moduledata
func ValidateModuleData(md *ModuleData, firstFuncEntry uint64) error {
	// Check if the first function's entry matches the minpc value
	if md.TextVA != firstFuncEntry {
		return fmt.Errorf("TextVA value %x does not match first function entry %x", md.TextVA, firstFuncEntry)
	}
	
	// Validate slice fields
	if md.Typelinks.Len > md.Typelinks.Capacity {
		return fmt.Errorf("invalid Typelinks slice: len %d > capacity %d", md.Typelinks.Len, md.Typelinks.Capacity)
	}
	
	if md.ITablinks.Len > md.ITablinks.Capacity {
		return fmt.Errorf("invalid ITablinks slice: len %d > capacity %d", md.ITablinks.Len, md.ITablinks.Capacity)
	}
	
	if md.LegacyTypes.Len > md.LegacyTypes.Capacity {
		return fmt.Errorf("invalid LegacyTypes slice: len %d > capacity %d", md.LegacyTypes.Len, md.LegacyTypes.Capacity)
	}
	
	if md.InitTasks.Len > md.InitTasks.Capacity {
		return fmt.Errorf("invalid InitTasks slice: len %d > capacity %d", md.InitTasks.Len, md.InitTasks.Capacity)
	}
	
	return nil
} 