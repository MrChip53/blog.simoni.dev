package helpers

import (
	"fmt"
	"time"
)

func FormatAsDateTime(t time.Time) string {
	year, month, day := t.Date()
	dateString := fmt.Sprintf("%d/%02d/%02d", year, month, day)

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		fmt.Println("Error loading time zone:", err)
		return dateString
	}

	chicagoTime := t.In(loc)

	timeFormat := "01/02/2006 3:04 PM"

	timeString := chicagoTime.Format(timeFormat)

	return timeString
}
