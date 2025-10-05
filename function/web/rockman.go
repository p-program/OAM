package web

type Rockman struct {
	Lazy struct{}
}

func (r *Rockman) SloveProblem(question string) (output interface{}, cost int64) {
	return "Go home and ask your mother", 0
}
