package pb

// Config for the ProgressBar
type Config struct {
	ShowPercent, ShowCounters        bool
	ShowSpeed, ShowTimeLeft, ShowBar bool
	ShowFinalTime                    bool
	NotPrint                         bool
	// ForceWidth if width is bigger than terminal width, will be ignored
	ForceWidth        bool
	DisableAutoUpdate bool
	AutoStat          bool
}

// NewConfig returns a defaulted config for most use cases will be appropriate
func NewConfig() Config {
	return Config{
		ShowPercent:       true,
		ShowCounters:      true,
		ShowBar:           true,
		ShowSpeed:         false,
		ShowTimeLeft:      true,
		ShowFinalTime:     true,
		DisableAutoUpdate: false,
		AutoStat:          false,
	}
}

// Build returns the resulting ProgressBar with all the Options from the config
func (cfg Config) Build(opts ...Option) *ProgressBar {
	return newProgressBar(cfg, opts...)
}
