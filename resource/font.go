package resource

type Font struct {
	Path string

	Size int
}

type FontID int

type FontRegistry struct {
	mapping map[FontID]Font
}

func (r *FontRegistry) Set(id FontID, info Font) {
	r.mapping[id] = info
}
