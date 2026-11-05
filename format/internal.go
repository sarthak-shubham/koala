package format

import (
	"encoding/binary"
)

func SerializeInternal(internal *InternalNode) ([]byte, error) {
	buf := make([]byte, PageSize)
	buf[0] = byte(PageTypeInternal)

	entriesCount := len(internal.Entries)
	binary.LittleEndian.PutUint16(buf[1:3], uint16(entriesCount))
	binary.LittleEndian.PutUint64(buf[3:11], internal.FirstChild)

	offset := 11
	for _, entry := range internal.Entries {
		keyLen := len(entry.Key)
		requiredBytes := 2 + keyLen + 8

		if offset+requiredBytes > PageSize {
			return nil, ErrPageDataTooLarge
		}

		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(keyLen))
		offset += 2

		copy(buf[offset:offset+keyLen], entry.Key)
		offset += keyLen

		binary.LittleEndian.PutUint64(buf[offset:offset+8], entry.PageID)
		offset += 8
	}

	return buf, nil
}

func DeserializeInternal(data []byte) (*InternalNode, error) {
	dataLen := len(data)
	if dataLen != PageSize {
		return nil, ErrInvalidPageSize
	}

	pageType := PageType(data[0])
	if pageType != PageTypeInternal {
		return nil, ErrCorruptedPage
	}

	entriesCount := int(binary.LittleEndian.Uint16(data[1:3]))
	firstChild := binary.LittleEndian.Uint64(data[3:11])

	entries := make([]InternalEntry, 0, entriesCount)
	offset := 11

	for i := 0; i < entriesCount; i++ {
		if offset+2 > PageSize {
			return nil, ErrCorruptedPage
		}

		keyLen := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2

		if offset+keyLen > PageSize {
			return nil, ErrCorruptedPage
		}

		key := make([]byte, keyLen)
		copy(key, data[offset:offset+keyLen])
		offset += keyLen

		if offset+8 > PageSize {
			return nil, ErrCorruptedPage
		}

		pageID := binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8

		entry := InternalEntry{
			Key:    key,
			PageID: pageID,
		}
		entries = append(entries, entry)
	}

	internal := &InternalNode{
		FirstChild: firstChild,
		Entries:    entries,
	}
	return internal, nil
}

