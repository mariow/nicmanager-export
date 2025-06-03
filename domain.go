package main

import (
	"time"
)

// Domain struct represents a domain entry from the API
type Domain struct {
	Name                 string `json:"name"`
	OrderStatus          string `json:"order_status"`
	OrderDateTime        string `json:"order_datetime"`
	RegistrationDateTime string `json:"registration_datetime"`
	DeleteDateTime       string `json:"delete_datetime"`
}

// parseAPIdate parses date strings from the Nicmanager API
func parseAPIdate(dateString string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", dateString)
}

// IsBelowCutoff filters for records without delete date or with delete date after cutoff
func (d *Domain) IsBelowCutoff(cutoffDate time.Time) bool {
	if d.DeleteDateTime != "" {
		parseDelDate, _ := parseAPIdate(d.DeleteDateTime)
		if parseDelDate.Unix() > cutoffDate.Unix() {
			return true
		}
	} else {
		return true
	}
	return false
}