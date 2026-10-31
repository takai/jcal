package holiday

import (
	"strings"
	"testing"
	"time"
)

// sjisHeader is the Shift_JIS header line of the official CSV
// ("国民の祝日・休日月日,国民の祝日・休日名称").
const sjisHeader = "\x8d\x91\x96\xaf\x82\xcc\x8f\x6a\x93\xfa\x81\x45\x8b\x78\x93\xfa\x8c\x8e\x93\xfa,\x8d\x91\x96\xaf\x82\xcc\x8f\x6a\x93\xfa\x81\x45\x8b\x78\x93\xfa\x96\xbc\x8f\xcc\r\n"

const sampleCSV = sjisHeader +
	"2026/1/1,\x8c\xb3\x93\xfa\r\n" +
	"2026/5/6,\x8b\x78\x93\xfa\r\n" +
	"2026/9/22,\x8b\x78\x93\xfa\r\n" +
	"2026/11/3,\x95\xb6\x89\xbb\x82\xcc\x93\xfa\r\n"

func TestParse(t *testing.T) {
	set, err := Parse(strings.NewReader(sampleCSV))
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 4 {
		t.Errorf("len = %d, want 4", len(set))
	}
	for _, d := range []struct {
		m time.Month
		d int
	}{{time.January, 1}, {time.May, 6}, {time.September, 22}, {time.November, 3}} {
		if !set.IsHoliday(2026, d.m, d.d) {
			t.Errorf("2026-%d-%d should be a holiday", d.m, d.d)
		}
	}
	if set.IsHoliday(2026, time.November, 4) {
		t.Error("2026-11-4 should not be a holiday")
	}
}

func TestParseRejectsDataWithoutDates(t *testing.T) {
	if _, err := Parse(strings.NewReader("<html>maintenance</html>\n")); err == nil {
		t.Error("expected error")
	}
}

func TestNilSet(t *testing.T) {
	var set Set
	if set.IsHoliday(2026, time.January, 1) {
		t.Error("nil set should report no holidays")
	}
}
