package buf

type RuneBuffer struct {
	data []rune
}

func (b *RuneBuffer) WriteAt(index int, s string) {
	runes := []rune(s)
	end := index + len(runes)

	if end > len(b.data) {
		newData := make([]rune, end)
		copy(newData, b.data)
		b.data = newData
	}

	copy(b.data[index:end], runes)
}

func (b *RuneBuffer) Rune(index int) (rune, bool) {
	if index >= 0 && index < len(b.data) {
		return b.data[index], true
	}
	return ' ', false
}

func (b *RuneBuffer) String() string {
	return string(b.data)
}
