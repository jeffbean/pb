// Package pb Simple console progress bars
package pb

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Version Current version
const Version = "1.0.13"

const (
	_defaultRefreshRate = time.Millisecond * 200
	// TODO: make format a type
	_defaultFormat = "[=>-]"
)

// Callback for custom output
// For example:
// bar.Callback = func(s string) {
//     mySuperPrint(s)
// }
//
type Callback func(out string)

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// ProgressBar is the struct containing the components to make the bar
type ProgressBar struct {
	current int64 // current must be first member of struct (https://code.google.com/p/go/issues/detail?id=5278)

	config Config

	goalValue int64
	Callback  Callback
	Units     Units
	Width     int

	// Default width for the time box.
	TimeBoxWidth int
	unitsWidth   int

	output       io.Writer
	outFormat    string
	unitFormater Formatter
	refreshRate  time.Duration

	finishOnce sync.Once //Guards isFinish
	finish     chan struct{}
	isFinish   bool

	startTime    time.Time
	startValue   int64
	currentValue int64

	prefix, postfix string

	mu        sync.Mutex
	lastPrint string

	BarStart string
	BarEnd   string
	Empty    string
	Current  string
	CurrentN string

	alwaysUpdate bool
}

// New Creates a new progress bar object
func New(total int, opts ...Option) *ProgressBar {
	opts = append(opts, WithGoalValue(total))
	return NewConfig().Build(opts...)
}

func newProgressBar(config Config, opts ...Option) *ProgressBar {
	pb := &ProgressBar{
		config:      config,
		Units:       NoUnit,
		output:      os.Stdout,
		refreshRate: _defaultRefreshRate,
		finish:      make(chan struct{}),
		outFormat:   _defaultFormat,
	}
	for _, opt := range opts {
		opt(pb)
	}
	return pb
}

// Start the progress twards the goal, this will start a timer to estimate the time.
func (pb *ProgressBar) Start() *ProgressBar {
	pb.startTime = time.Now()
	pb.startValue = atomic.LoadInt64(&pb.current)
	if pb.goalValue == 0 {
		pb.config.ShowTimeLeft = false
		pb.config.ShowPercent = false
		pb.config.AutoStat = false
	}
	if !pb.config.DisableAutoUpdate {
		pb.Update() // Initial printing of the bar before running the bar refresher.
		go pb.refresher()
	}
	return pb
}

// Format set custom format for bar
// Example: bar.Format("[=>_]")
// Example: bar.Format("[\x00=\x00>\x00-\x00]") // \x00 is the delimiter
func (pb *ProgressBar) Format(format string) *ProgressBar {
	var formatEntries []string
	if utf8.RuneCountInString(format) == 5 {
		formatEntries = strings.Split(format, "")
	} else {
		formatEntries = strings.Split(format, "\x00")
	}
	if len(formatEntries) == 5 {
		pb.BarStart = formatEntries[0]
		pb.BarEnd = formatEntries[4]
		pb.Empty = formatEntries[3]
		pb.Current = formatEntries[1]
		pb.CurrentN = formatEntries[2]
	}
	return pb
}

// Increment current value
func (pb *ProgressBar) Increment() {
	pb.Add(1)
}

// Get current value
func (pb *ProgressBar) Get() int {
	return int(atomic.LoadInt64(&pb.current))
}

// Set current value
func (pb *ProgressBar) Set(current int) {
	atomic.StoreInt64(&pb.current, int64(current))
}

// Add to current value
func (pb *ProgressBar) Add(add int) {
	atomic.AddInt64(&pb.current, int64(add))
}

// Finish stops the refresh updates to the output writer.
func (pb *ProgressBar) Finish() {
	//Protect multiple calls
	pb.finishOnce.Do(func() {
		close(pb.finish)
		pb.write(atomic.LoadInt64(&pb.current))
		pb.mu.Lock()
		defer pb.mu.Unlock()
		switch {
		case pb.output != nil:
			fmt.Fprintln(pb.output)
		case !pb.config.NotPrint:
			fmt.Println()
		}
		pb.isFinish = true
	})
}

// IsFinished return boolean
func (pb *ProgressBar) IsFinished() bool {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.isFinish
}

// FinishPrint ends printing and writes a final string on a new line
func (pb *ProgressBar) FinishPrint(str string) {
	pb.Finish()
	if pb.output != nil {
		fmt.Fprintln(pb.output, str)
	} else {
		fmt.Println(str)
	}
}

// implement io.Writer
func (pb *ProgressBar) Write(p []byte) (n int, err error) {
	n = len(p)
	pb.Add(n)
	return
}

// implement io.Reader
func (pb *ProgressBar) Read(p []byte) (n int, err error) {
	n = len(p)
	pb.Add(n)
	return
}

// NewProxyReader Create new proxy reader over bar
// Takes io.Reader or io.ReadCloser
func (pb *ProgressBar) NewProxyReader(r io.Reader) *Reader {
	return &Reader{r, pb}
}

func (pb *ProgressBar) write(current int64) {
	width := pb.GetWidth()

	var percentBox, countersBox, timeLeftBox, speedBox, barBox, end, out string

	// percents
	if pb.config.ShowPercent {
		var percent float64
		if pb.goalValue > 0 {
			percent = float64(current) / (float64(pb.goalValue) / float64(100))
		} else {
			percent = float64(current) / float64(100)
		}
		percentBox = fmt.Sprintf(" %6.02f%%", percent)
	}

	// counters
	if pb.config.ShowCounters {
		current := NewFormatter(current, pb)
		if pb.goalValue > 0 {
			total := NewFormatter(pb.goalValue, pb)
			countersBox = fmt.Sprintf(" %s / %s ", current, total)
		} else {
			countersBox = fmt.Sprintf(" %s / ? ", current)
		}
	}

	// time left
	fromStart := time.Now().Sub(pb.startTime)
	currentFromStart := current - pb.startValue
	select {
	case <-pb.finish:
		if pb.config.ShowFinalTime {
			var left time.Duration
			left = (fromStart / time.Second) * time.Second
			timeLeftBox = fmt.Sprintf(" %s", left.String())
		}
	default:
		if pb.config.ShowTimeLeft && currentFromStart > 0 {
			perEntry := fromStart / time.Duration(currentFromStart)
			var left time.Duration
			if pb.goalValue > 0 {
				left = time.Duration(pb.goalValue-currentFromStart) * perEntry
				left = (left / time.Second) * time.Second
			} else {
				left = time.Duration(currentFromStart) * perEntry
				left = (left / time.Second) * time.Second
			}
			timeLeft := NewFormatter(int64(left), pb)
			timeLeftBox = fmt.Sprintf(" %s", timeLeft)
		}
	}

	if len(timeLeftBox) < pb.TimeBoxWidth {
		timeLeftBox = fmt.Sprintf("%s%s", strings.Repeat(" ", pb.TimeBoxWidth-len(timeLeftBox)), timeLeftBox)
	}

	// speed
	if pb.config.ShowSpeed && currentFromStart > 0 {
		fromStart := time.Now().Sub(pb.startTime)
		speed := float64(currentFromStart) / (float64(fromStart) / float64(time.Second))
		speedBox = " " + NewFormatter(int64(speed), pb, PerSec()).String()
	}

	barWidth := escapeAwareRuneCountInString(countersBox + pb.BarStart + pb.BarEnd + percentBox + timeLeftBox + speedBox + pb.prefix + pb.postfix)
	// bar
	if pb.config.ShowBar {
		size := width - barWidth
		if size > 0 {
			if pb.goalValue > 0 {
				curCount := int(math.Ceil((float64(current) / float64(pb.goalValue)) * float64(size)))
				emptCount := size - curCount
				barBox = pb.BarStart
				if emptCount < 0 {
					emptCount = 0
				}
				if curCount > size {
					curCount = size
				}
				if emptCount <= 0 {
					barBox += strings.Repeat(pb.Current, curCount)
				} else if curCount > 0 {
					barBox += strings.Repeat(pb.Current, curCount-1) + pb.CurrentN
				}
				barBox += strings.Repeat(pb.Empty, emptCount) + pb.BarEnd
			} else {
				barBox = pb.BarStart
				pos := size - int(current)%int(size)
				if pos-1 > 0 {
					barBox += strings.Repeat(pb.Empty, pos-1)
				}
				barBox += pb.Current
				if size-pos-1 > 0 {
					barBox += strings.Repeat(pb.Empty, size-pos-1)
				}
				barBox += pb.BarEnd
			}
		}
	}

	// check len
	out = pb.prefix + countersBox + barBox + percentBox + speedBox + timeLeftBox + pb.postfix
	if cl := escapeAwareRuneCountInString(out); cl < width {
		end = strings.Repeat(" ", width-cl)
	}

	// and print!
	pb.mu.Lock()
	pb.lastPrint = out + end
	isFinish := pb.isFinish
	pb.mu.Unlock()
	switch {
	case isFinish:
		return
	case pb.output != nil:
		fmt.Fprint(pb.output, "\r"+out+end)
	case pb.Callback != nil:
		pb.Callback(out + end)
	case !pb.config.NotPrint:
		fmt.Print("\r" + out + end)
	}
}

// GetTerminalWidth - returns terminal width for all platforms.
func GetTerminalWidth() (int, error) {
	return terminalWidth()
}

// GetWidth returns the witdh of the bar
func (pb *ProgressBar) GetWidth() int {
	if pb.config.ForceWidth {
		return pb.Width
	}

	width := pb.Width
	termWidth, _ := terminalWidth()
	if width == 0 || termWidth <= width {
		width = termWidth
	}

	return width
}

//Update Write the current state of the progressbar
func (pb *ProgressBar) Update() {
	c := atomic.LoadInt64(&pb.current)
	if pb.alwaysUpdate || pb.currentValue != c {
		pb.write(c)
		pb.currentValue = c
	}
	fmt.Printf("autostat set: %v\n", pb.config.AutoStat)
	if pb.config.AutoStat {
		if c == 0 {
			pb.startTime = time.Now()
			pb.startValue = 0
		} else if c >= pb.goalValue && pb.isFinish != true {
			pb.Finish()
		}
	}
}

// String return the last bar print
func (pb *ProgressBar) String() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.lastPrint
}

// Internal loop for refreshing the progressbar
func (pb *ProgressBar) refresher() {
	for {
		select {
		case <-pb.finish:
			return
		case <-time.After(pb.refreshRate):
			pb.Update()
		}
	}
}
