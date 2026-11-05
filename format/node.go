package format

const (
	PageSize = 4096

	MagicNumber uint32 = 0x4B4C4442

	FormatVersion uint32 = 1
)

type PageType byte

const (
	PageTypeMeta     PageType = 0x00
	PageTypeLeaf     PageType = 0x01
	PageTypeInternal PageType = 0x02
)

type MetaPage struct {
	MagicNumber uint32
	Version     uint32
	PageSize    uint32
	RootPageID  uint64
	PageCount   uint64
}

type LeafEntry struct {
	Key   []byte
	Value []byte
}

type LeafNode struct {
	PrevPageID uint64
	NextPageID uint64
	Entries    []LeafEntry
}

type InternalEntry struct {
	Key    []byte
	PageID uint64
}

type InternalNode struct {
	FirstChild uint64
	Entries    []InternalEntry
}

type Node interface {
	Type() PageType
}

func (l *LeafNode) Type() PageType     { return PageTypeLeaf }
func (i *InternalNode) Type() PageType { return PageTypeInternal }

