package pb

import (
	"fmt"
	"time"
)

// Units represents different types of untis supported by the Progress bar
type Units int

const (
	// NoUnit are default units, they represent a simple value and are not formatted at all.
	NoUnit Units = iota
	// DataSizeUnit units are formatted in a human readable way (b, Bb, Mb, ...)
	DataSizeUnit
	// DurationUnit units are formatted in a human readable way (3h14m15s)
	DurationUnit
)

const (
	KiB = 1024
	MiB = 1048576
	GiB = 1073741824
	TiB = 1099511627776
)

// FormatOption are option closure for construction on  the unti formatter
type FormatOption func(f *Formatter)

// Formatter is a struct to print on the progress bar without dealing with unit conversions
type Formatter struct {
	n      int64
	unit   Units
	width  int
	perSec bool
}

// NewFormatter returns a unit formatter
func NewFormatter(i int64, pb *ProgressBar, opts ...FormatOption) *Formatter {
	f := &Formatter{
		n:     i,
		unit:  pb.Units,
		width: pb.UnitsWidth,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// PerSec adds an extra "/s" in the string formatting
func PerSec() FormatOption {
	return func(f *Formatter) {
		f.perSec = true
	}
}

func (f *Formatter) String() string {
	var out string
	switch f.unit {
	case DataSizeUnit:
		out = formatBytes(f.n)
	case DurationUnit:
		out = formatDuration(f.n)
	default:
		out = fmt.Sprintf(fmt.Sprintf("%%%dd", f.width), f.n)
	}
	if f.perSec {
		out += "/s"
	}
	return out
}

// Convert bytes to human readable string. Like a 2 MiB, 64.2 KiB, 52 B
func formatBytes(i int64) (result string) {
	switch {
	case i >= TiB:
		result = fmt.Sprintf("%.02f TiB", float64(i)/TiB)
	case i >= GiB:
		result = fmt.Sprintf("%.02f GiB", float64(i)/GiB)
	case i >= MiB:
		result = fmt.Sprintf("%.02f MiB", float64(i)/MiB)
	case i >= KiB:
		result = fmt.Sprintf("%.02f KiB", float64(i)/KiB)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	return
}

func formatDuration(n int64) (result string) {
	d := time.Duration(n)
	if d > time.Hour*24 {
		result = fmt.Sprintf("%dd", d/24/time.Hour)
		d -= (d / time.Hour / 24) * (time.Hour * 24)
	}
	result = fmt.Sprintf("%s%v", result, d)
	return
}
