package crw

import "time"

type (
	VozConfig struct {
		TheadUrl    string
		NumWorker   int
		TimeToWrite time.Duration
	}
)
