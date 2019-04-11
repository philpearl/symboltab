package main

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/loov/hrtime"
	"github.com/philpearl/symboltab"
)

const count = 1e7

func main() {
	b := hrtime.NewBenchmarkTSC(count)

	symbols := make([]string, count)
	for i := range symbols {
		symbols[i] = strconv.Itoa(i)
	}

	st := symboltab.New(0)

	runtime.GC()

	for i := 0; b.Next(); i++ {
		if i >= count {
			i = 0
		}
		t := hrtime.TSC()
		st.StringToSequence(symbols[i], true)
		st.StringToSequence(symbols[i], true)
		dur := hrtime.TSC() - t
		if dur.ApproxDuration() > time.Millisecond*100 {
			// When we grow the table to larger sizes we see slow performance. It seems that just allocating
			// these very big slices takes > 100ms, presumably because they are zeroed
			fmt.Printf("big number at %d\n", i)
		}
	}

	opts := hrtime.HistogramOptions{
		BinCount:        20,
		NiceRange:       true,
		ClampMaximum:    0,
		ClampPercentile: 0.999999,
	}
	fmt.Println(hrtime.NewDurationHistogram(b.Laps(), &opts))
}
