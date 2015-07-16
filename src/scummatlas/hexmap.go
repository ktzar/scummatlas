package scummatlas

type HexMapSection struct {
	Start       int
	Length      int
	Type        string
	Description string
}

func (s HexMapSection) IncludesOffset(offset int) bool {
	return offset > s.Start && offset < s.Start+s.Length
}

type HexMap struct {
	data     []byte
	sections []HexMapSection
}

func (h HexMap) Data() []byte {
	return h.data
}

func (h HexMap) Sections() []HexMapSection {
	return h.sections
}

func (h *HexMap) AddSection(section HexMapSection) {
	h.sections = append(h.sections, section)
}
