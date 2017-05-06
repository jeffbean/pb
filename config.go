package pb

// Config holds some options for the progress bar
type Config struct {
	ShowPercent, ShowCounters        bool
	ShowSpeed, ShowTimeLeft, ShowBar bool
	ShowFinalTime                    bool
	NotPrint                         bool
	Units                            Units
}
