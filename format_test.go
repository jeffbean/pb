package pb

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func Test_DefaultsToInteger(t *testing.T) {
	value := int64(1000)
	expected := strconv.Itoa(int(value))
	actual := NewFormatter(value, New(1)).String()

	if actual != expected {
		t.Errorf(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_CanFormatAsInteger(t *testing.T) {
	value := int64(1000)
	expected := strconv.Itoa(int(value))
	actual := NewFormatter(value, New(1, WithUnits(NoUnit))).String()

	if actual != expected {
		t.Errorf(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_CanFormatAsBytes(t *testing.T) {
	inputs := []struct {
		v int64
		e string
	}{
		{v: 1000, e: "1000 B"},
		{v: 1024, e: "1.00 KiB"},
		{v: 3*miB + 140*kiB, e: "3.14 MiB"},
		{v: 2 * giB, e: "2.00 GiB"},
		{v: 2048 * giB, e: "2.00 TiB"},
	}

	for _, input := range inputs {
		actual := NewFormatter(input.v, New(1, WithUnits(DataSizeUnit))).String()
		if actual != input.e {
			t.Errorf(fmt.Sprintf("Expected {%s} was {%s}", input.e, actual))
		}
	}
}

func Test_CanFormatDuration(t *testing.T) {
	value := 10 * time.Minute
	expected := "10m0s"
	actual := NewFormatter(int64(value), New(1, WithUnits(DurationUnit))).String()
	if actual != expected {
		t.Errorf(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}

func Test_DefaultUnitsWidth(t *testing.T) {
	value := 10
	expected := "     10"
	actual := NewFormatter(int64(value), New(20, WithWidth(10))).String()
	if actual != expected {
		t.Errorf(fmt.Sprintf("Expected {%s} was {%s}", expected, actual))
	}
}
