package proto

type Bounds struct {
	MinX float32
	MinY float32
	MaxX float32
	MaxY float32
}

type rectBottomLeft struct {
	x      float32
	y      float32
	width  float32
	height float32
}

func calculateRectBottomLeftBounds(size *Size, position *Point, anchorPoint *Point) *rectBottomLeft {
	offsetX := anchorPoint.X * size.Width
	offsetYFromBottom := (1.0 - anchorPoint.Y) * size.Height
	bottomLeftX := position.X - offsetX
	bottomLeftY := position.Y - offsetYFromBottom
	return &rectBottomLeft{
		x:      bottomLeftX,
		y:      bottomLeftY,
		width:  size.Width,
		height: size.Height,
	}
}

func (r *rectBottomLeft) toBounds() *Bounds {
	return &Bounds{
		MinX: r.x,
		MinY: r.y,
		MaxX: r.x + r.width,
		MaxY: r.y + r.height,
	}
}

func NewBounds(s *Size, p *Point) *Bounds {
	return calculateRectBottomLeftBounds(s, p, &Point{X: 0.5, Y: 0.5}).toBounds()
}

func (b *Bounds) Width() float32 {
	return b.MaxX - b.MinX
}

func (b *Bounds) Height() float32 {
	return b.MaxY - b.MinY
}
