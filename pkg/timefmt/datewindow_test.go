package timefmt

import (
	"errors"
	"testing"
	"time"
)

func TestParseDateWindow(t *testing.T) {
	tests := []struct {
		name      string
		since     string
		until     string
		wantSince time.Time
		wantUntil time.Time
		wantErr   bool
	}{
		{
			name:      "两侧为空表示不限",
			wantSince: time.Time{},
			wantUntil: time.Time{},
		},
		{
			name:      "仅since取当日零点",
			since:     "2024-03-05",
			wantSince: time.Date(2024, 3, 5, 0, 0, 0, 0, time.Local),
			wantUntil: time.Time{},
		},
		{
			name:      "仅until取当日末尾",
			until:     "2024-03-05",
			wantSince: time.Time{},
			wantUntil: time.Date(2024, 3, 5, 23, 59, 59, 999999999, time.Local),
		},
		{
			name:      "双侧窗口",
			since:     "2024-01-01",
			until:     "2024-12-31",
			wantSince: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			wantUntil: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.Local),
		},
		{name: "非法since缺零填充", since: "2024-3-5", wantErr: true},
		{name: "非法until乱码", until: "not-a-date", wantErr: true},
		{name: "非法until分隔符", until: "2024/03/05", wantErr: true},
		{name: "非法since月份越界", since: "2024-13-01", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sinceTime, untilTime, err := ParseDateWindow(tt.since, tt.until)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseDateWindow(%q, %q) 期望报错，得到 nil", tt.since, tt.until)
				}
				var dwe *DateWindowError
				if !errors.As(err, &dwe) {
					t.Fatalf("期望 *DateWindowError，得到 %T: %v", err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDateWindow(%q, %q) 报错: %v", tt.since, tt.until, err)
			}
			if !sinceTime.Equal(tt.wantSince) {
				t.Errorf("sinceTime = %v, 期望 %v", sinceTime, tt.wantSince)
			}
			if !untilTime.Equal(tt.wantUntil) {
				t.Errorf("untilTime = %v, 期望 %v", untilTime, tt.wantUntil)
			}
		})
	}
}

func TestInDateWindow(t *testing.T) {
	since := time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local)
	until := time.Date(2024, 1, 20, 23, 59, 59, 999999999, time.Local)

	tests := []struct {
		name   string
		t      time.Time
		since  time.Time
		until  time.Time
		wantIn bool
	}{
		{name: "窗口起点当日中午包含", t: time.Date(2024, 1, 10, 12, 0, 0, 0, time.Local), since: since, until: until, wantIn: true},
		{name: "恰为since包含", t: since, since: since, until: until, wantIn: true},
		{name: "早于since排除", t: time.Date(2024, 1, 9, 23, 59, 59, 0, time.Local), since: since, until: until, wantIn: false},
		{name: "恰为until包含", t: until, since: since, until: until, wantIn: true},
		{name: "晚于until排除", t: time.Date(2024, 1, 21, 0, 0, 0, 0, time.Local), since: since, until: until, wantIn: false},
		{name: "无窗口恒包含", t: time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local), wantIn: true},
		{name: "仅until时早期提交包含", t: time.Date(2020, 6, 1, 0, 0, 0, 0, time.Local), until: until, wantIn: true},
		{name: "仅since时晚期提交包含", t: time.Date(2030, 6, 1, 0, 0, 0, 0, time.Local), since: since, wantIn: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InDateWindow(tt.t, tt.since, tt.until); got != tt.wantIn {
				t.Errorf("InDateWindow(%v, %v, %v) = %v, 期望 %v", tt.t, tt.since, tt.until, got, tt.wantIn)
			}
		})
	}
}
