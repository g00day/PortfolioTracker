package utils

import (
	"time"
)

/* This file stores the function that
converts time data, defaultly time.Time

to RFC3339 formating that is used to store timestamps in postgreSQL
*/

func ConvertTime(t time.Time) string {

	timestamp := t.Format(time.RFC3339)

	return timestamp

}
