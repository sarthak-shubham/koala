package format

import (
	"encoding/binary"
)

func SerializeLeaf(leaf *LeafNode) ([]byte, error) {
	buf := make([]byte, PageSize)
	buf[0] = byte(PageTypeLeaf)

	entriesCount := len(leaf.Entries)
	binary.LittleEndian.PutUint16(buf[1:3], uint16(entriesCount))
	binary.LittleEndian.PutUint64(buf[3:11], leaf.PrevPageID)
	binary.LittleEndian.PutUint64(buf[11:19], leaf.NextPageID)

	offset := 19
	for _, entry := range leaf.Entries {
		keyLen := len(entry.Key)
		valLen := len(entry.Value)
		requiredBytes := 2 + keyLen + 2 + valLen

		if offset+requiredBytes > PageSize {
			return nil, ErrPageDataTooLarge
		}

		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(keyLen))
		offset += 2

		copy(buf[offset:offset+keyLen], entry.Key)
		offset += keyLen

		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(valLen))
		offset += 2

		copy(buf[offset:offset+valLen], entry.Value)
		offset += valLen
	}

	return buf, nil
}

func DeserializeLeaf(data []byte) (*LeafNode, error) {
	dataLen := len(data)
	if dataLen != PageSize {
		return nil, ErrInvalidPageSize
	}

	pageType := PageType(data[0])
	if pageType != PageTypeLeaf {
		return nil, ErrCorruptedPage
	}

	entriesCount := int(binary.LittleEndian.Uint16(data[1:3]))
	prevPageID := binary.LittleEndian.Uint64(data[3:11])
	nextPageID := binary.LittleEndian.Uint64(data[11:19])

	entries := make([]LeafEntry, 0, entriesCount)
	offset := 19

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

		if offset+2 > PageSize {
			return nil, ErrCorruptedPage
		}

		valLen := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2

		if offset+valLen > PageSize {
			return nil, ErrCorruptedPage
		}

		value := make([]byte, valLen)
		copy(value, data[offset:offset+valLen])
		offset += valLen

		entry := LeafEntry{
			Key:   key,
			Value: value,
		}
		entries = append(entries, entry)
	}

	leaf := &LeafNode{
		PrevPageID: prevPageID,
		NextPageID: nextPageID,
		Entries:    entries,
	}
	return leaf, nil
}


