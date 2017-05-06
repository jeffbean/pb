package pb_test

import (
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

func Example() {
	count := 5000
	cfg := pb.Config{
		ShowPercent:  true,
		ShowBar:      true,
		ShowCounters: true,
		ShowTimeLeft: true,
	}
	bar := cfg.Build(pb.WithGoalValue(count))
	// and start
	bar.Start()
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}
	bar.FinishPrint("The End!")
}
