package web

import "time"

type AI interface {
	SloveProblem(question string) (output interface{}, cost time.Duration)
}
