package calendar

import (
	"fmt"
	"time"
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func FromTime(t time.Time, location *time.Location) Date {
	if location != nil {
		t = t.In(location)
	}
	return GregorianToJalali(t.Year(), int(t.Month()), t.Day())
}

func (d Date) YearMonth() string {
	return fmt.Sprintf("%04d-%02d", d.Year, d.Month)
}

func GregorianToJalali(gy, gm, gd int) Date {
	gdm := [...]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	if gy > 1600 {
		gy -= 1600
	} else {
		gy -= 621
	}

	gy2 := gy
	if gm > 2 {
		gy2++
	}

	days := 365*gy + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 - 80 + gd + gdm[gm-1]
	jy := 979 + 33*(days/12053)
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461

	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}

	var jm, jd int
	if days < 186 {
		jm = 1 + days/31
		jd = 1 + days%31
	} else {
		jm = 7 + (days-186)/30
		jd = 1 + (days-186)%30
	}

	return Date{Year: jy, Month: jm, Day: jd}
}
