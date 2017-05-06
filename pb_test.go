package pb

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/stretchr/testify/assert"
)

func Test_IncrementAddsOne(t *testing.T) {
	tests := []struct {
		goal     int
		incCount int
		want     int
	}{
		{5000, 20, 20},
		{10, 25, 25},
		{0, 10, 10},
		{-5, 2, 2},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Increment count %v", tt.incCount), func(t *testing.T) {
			bar := New(tt.goal)
			for i := 0; i < tt.incCount; i++ {
				bar.Increment()
			}
			assert.Equal(t, tt.want, bar.Get(), "Increment should be adding by 1")
		})
	}
}

func TestProgressSet(t *testing.T) {
	tests := []struct {
		goal int
		set  int
		want int
	}{
		{5000, 20, 20},
		{10, 25, 25},
		{0, 10, 10},
		{-5, 2, 2},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Setting %v", tt.set), func(t *testing.T) {
			bar := New(tt.goal)
			bar.Set(tt.set)
			assert.Equal(t, tt.want, bar.Get(), "Increment should be adding by 1")
		})
	}
}

func TestProgressAdd(t *testing.T) {
	tests := []struct {
		goal     int
		addCount int
		addValue int
		want     int
	}{
		{5000, 5, 2, 10},
		{10, 100, 3, 300},
		{1800, 24, 75, 1800},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Setting %v", tt.addCount*tt.addValue), func(t *testing.T) {
			bar := New(tt.goal)
			bar.Start()
			for i := 0; i < tt.addCount; i++ {
				bar.Add(tt.addValue)
			}
			assert.Equal(t, tt.want, bar.Get(), "Increment should be adding by 1")
		})
	}
}

func TestWriteRace(t *testing.T) {
	outBuffer := &bytes.Buffer{}
	totalCount := 20
	bar := New(totalCount)
	bar.output = outBuffer
	bar.Start()
	var wg sync.WaitGroup
	for i := 0; i < totalCount; i++ {
		wg.Add(1)
		go func() {
			bar.Increment()
			time.Sleep(250 * time.Millisecond)
			wg.Done()
		}()
	}
	wg.Wait()
	bar.Finish()
}

func Test_MultipleFinish(t *testing.T) {
	bar := New(5000)
	bar.Add(2000)
	bar.Finish()
	bar.Finish()

	assert.Equal(t, 2000, bar.Get())
	assert.Equal(t, int64(5000), bar.goalValue)
}

func Test_Format(t *testing.T) {
	bar := New(5000).Format(strings.Join([]string{
		color.GreenString("["),
		color.New(color.BgGreen).SprintFunc()("o"),
		color.New(color.BgHiGreen).SprintFunc()("o"),
		color.New(color.BgRed).SprintFunc()("o"),
		color.GreenString("]"),
	}, "\x00"))
	w := colorable.NewColorableStdout()
	bar.Callback = func(out string) {
		w.Write([]byte(out))
	}
	bar.Add(2000)
	bar.Finish()
	bar.Finish()
}

func Test_AutoStat(t *testing.T) {
	tests := []struct {
		goal      int
		workCount int
		addValue  int
		want      int
	}{
		{5000, 5, 2, 10},
		{10, 100, 3, 300},
		{1800, 24, 75, 1800},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Setting %v", tt.workCount*tt.addValue), func(t *testing.T) {
			bar := Config{
				AutoStat: true,
			}.Build(WithGoalValue(5))
			bar.Start()

			for i := 0; i < tt.workCount; i++ {
				bar.Add(tt.addValue)
			}
			assert.Equal(t, tt.want, bar.Get(), "Increment should be adding by 1")
		})
	}
}

func Test_Finish_PrintNewline(t *testing.T) {
	buf := &bytes.Buffer{}
	bar := New(5, WithOutput(buf))

	bar.output = buf
	bar.Finish()

	expected := "\n"
	actual := buf.String()
	//Finish should write newline to bar.Output
	if !strings.HasSuffix(actual, expected) {
		t.Errorf("Expected %q to have suffix %q", expected, actual)
	}
}

func Test_FinishPrint(t *testing.T) {
	bar := New(5)
	buf := &bytes.Buffer{}
	bar.output = buf
	bar.FinishPrint("foo")

	expected := "foo\n"
	actual := buf.String()
	//FinishPrint should write to bar.Output
	if !strings.HasSuffix(actual, expected) {
		t.Errorf("Expected %q to have suffix %q", expected, actual)
	}
}
