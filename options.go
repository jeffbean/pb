package pb

import (
	"io"
	"time"
)

// Option is a Progress bar option to create it with
type Option func(*ProgressBar)

// WithGoalValue sets the goal number for the bar to aim twards
func WithGoalValue(goal int) Option {
	return func(pb *ProgressBar) {
		pb.goalValue = int64(goal)
	}
}

// WithFormat sets the output string format
// Needs to conform to a 5 char format structure eg, "[=>-]"
func WithFormat(format string) Option {
	return func(pb *ProgressBar) {
		pb.outFormat = format
	}
}

// WithPrefix sets the prefix string
func WithPrefix(prefix string) Option {
	return func(pb *ProgressBar) {
		pb.prefix = prefix
	}
}

// WithPostfix sets the postfix string
func WithPostfix(postfix string) Option {
	return func(pb *ProgressBar) {
		pb.postfix = postfix
	}
}

// WithOutput sets the bars output writer
func WithOutput(w io.Writer) Option {
	return func(pb *ProgressBar) {
		pb.output = w
	}
}

// WithRefreshRate configures the  refresh rate
func WithRefreshRate(rate time.Duration) Option {
	return func(pb *ProgressBar) {
		pb.refreshRate = rate
	}
}

// WithUnits Set the unit of measure
// bar.SetUnits(U_NO) - by default
// bar.SetUnits(U_BYTES) - for Mb, Kb, etc
func WithUnits(units Units) Option {
	return func(pb *ProgressBar) {
		pb.Units = units
	}
}

// WithWidth sets the bar width
func WithWidth(width int) Option {
	return func(pb *ProgressBar) {
		pb.Width = width
	}
}

// WithUnitsWidth sets the bar width
func WithUnitsWidth(width int) Option {
	return func(pb *ProgressBar) {
		pb.unitsWidth = width
	}
}
