package proto

func (b *Bounds) Width() float32 {
	return b.MaxX - b.MinX
}

func (b *Bounds) Height() float32 {
	return b.MaxY - b.MinY
}
