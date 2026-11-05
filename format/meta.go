package format

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidPageSize    = errors.New("data size does not match page size")
	ErrInvalidMagicNumber = errors.New("invalid magic number")
	ErrUnknownPageType    = errors.New("unknown page type")
	ErrPageDataTooLarge   = errors.New("page data exceeds page size")
	ErrCorruptedPage      = errors.New("corrupted page data")
)

func SerializeMeta(meta *MetaPage) ([]byte, error) {
	buf := make([]byte, PageSize)

	binary.LittleEndian.PutUint32(buf[0:4], meta.MagicNumber)
	binary.LittleEndian.PutUint32(buf[4:8], meta.Version)
	binary.LittleEndian.PutUint32(buf[8:12], meta.PageSize)
	binary.LittleEndian.PutUint64(buf[12:20], meta.RootPageID)
	binary.LittleEndian.PutUint64(buf[20:28], meta.PageCount)

	return buf, nil
}

func DeserializeMeta(data []byte) (*MetaPage, error) {
	dataLen := len(data)
	if dataLen != PageSize {
		return nil, ErrInvalidPageSize
	}

	magicNumber := binary.LittleEndian.Uint32(data[0:4])
	if magicNumber != MagicNumber {
		return nil, ErrInvalidMagicNumber
	}

	version := binary.LittleEndian.Uint32(data[4:8])
	pageSize := binary.LittleEndian.Uint32(data[8:12])
	rootPageID := binary.LittleEndian.Uint64(data[12:20])
	pageCount := binary.LittleEndian.Uint64(data[20:28])

	meta := &MetaPage{
		MagicNumber: magicNumber,
		Version:     version,
		PageSize:    pageSize,
		RootPageID:  rootPageID,
		PageCount:   pageCount,
	}
	return meta, nil
}

func DeserializeNode(data []byte) (Node, error) {
	dataLen := len(data)
	if dataLen != PageSize {
		return nil, ErrInvalidPageSize
	}

	pageType := PageType(data[0])

	if pageType == PageTypeLeaf {
		leaf, err := DeserializeLeaf(data)
		if err != nil {
			return nil, err
		}
		return leaf, nil
	}

	if pageType == PageTypeInternal {
		internal, err := DeserializeInternal(data)
		if err != nil {
			return nil, err
		}
		return internal, nil
	}

	return nil, ErrUnknownPageType
}

