package pager

import (
	"errors"
	"os"

	"github.com/sarthak-shubham/koala/format"
)

var (
	ErrInvalidPageSize = errors.New("data size does not match page size")
)

type Pager struct {
	file      *os.File
	pageCount uint64
}

func NewPager(file *os.File) *Pager {
	info, _ := file.Stat()
	pageCount := uint64(info.Size()) / format.PageSize
	pager := &Pager{
		file:      file,
		pageCount: pageCount,
	}
	return pager
}

func (p *Pager) PageOffset(pageID uint64) int64 {
	offset := int64(pageID * format.PageSize)
	return offset
}

func (p *Pager) PageCount() uint64 {
	count := p.pageCount
	return count
}

func (p *Pager) AllocatePage() (uint64, error) {
	pageID := p.pageCount
	empty := make([]byte, format.PageSize)
	err := p.WritePage(pageID, empty)
	if err != nil {
		return 0, err
	}
	p.pageCount++
	return pageID, nil
}

func (p *Pager) ReadPage(pageID uint64) ([]byte, error) {
	buf := make([]byte, format.PageSize)
	offset := p.PageOffset(pageID)
	_, err := p.file.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (p *Pager) WritePage(pageID uint64, data []byte) error {
	dataLen := len(data)
	if dataLen != format.PageSize {
		return ErrInvalidPageSize
	}
	offset := p.PageOffset(pageID)
	_, err := p.file.WriteAt(data, offset)
	if err != nil {
		return err
	}
	if pageID >= p.pageCount {
		p.pageCount = pageID + 1
	}
	return nil
}

func (p *Pager) Sync() error {
	err := p.file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (p *Pager) Close() error {
	err := p.file.Close()
	if err != nil {
		return err
	}
	return nil
}
