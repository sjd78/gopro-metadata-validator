package main

import (
	"fmt"
	"time"
)

// getTimezoneFromCoordinates returns timezone location from GPS coordinates
// Uses simplified zone mapping - not perfect but good enough for most cases
// For more accuracy, consider using github.com/evanoberholster/timezoneLookup
func getTimezoneFromCoordinates(lat, lon float64) *time.Location {
	// UTC offset estimation: longitude / 15 degrees per hour
	hourOffset := int(lon / 15.0)

	// Handle special cases for major timezone boundaries
	// This is a simplified approach - production code would use timezone boundary database

	// North America - United States
	if lat >= 25 && lat <= 50 && lon >= -125 && lon <= -65 {
		if lon > -75 {
			return loadLocation("America/New_York") // EST/EDT (-05:00/-04:00)
		} else if lon > -90 {
			return loadLocation("America/Chicago") // CST/CDT (-06:00/-05:00)
		} else if lon > -105 {
			return loadLocation("America/Denver") // MST/MDT (-07:00/-06:00)
		} else {
			return loadLocation("America/Los_Angeles") // PST/PDT (-08:00/-07:00)
		}
	}

	// North America - Canada
	if lat >= 42 && lat <= 70 && lon >= -141 && lon <= -52 {
		if lon > -60 {
			return loadLocation("America/Halifax") // AST/ADT (-04:00/-03:00)
		} else if lon > -70 {
			return loadLocation("America/Toronto") // EST/EDT
		} else if lon > -90 {
			return loadLocation("America/Winnipeg") // CST/CDT
		} else if lon > -115 {
			return loadLocation("America/Edmonton") // MST/MDT
		} else {
			return loadLocation("America/Vancouver") // PST/PDT
		}
	}

	// Europe
	if lat >= 35 && lat <= 70 && lon >= -10 && lon <= 40 {
		if lon < 15 {
			return loadLocation("Europe/London") // GMT/BST (+00:00/+01:00)
		} else {
			return loadLocation("Europe/Paris") // CET/CEST (+01:00/+02:00)
		}
	}

	// Australia
	if lat >= -45 && lat <= -10 && lon >= 113 && lon <= 154 {
		if lon < 130 {
			return loadLocation("Australia/Perth") // AWST (+08:00)
		} else if lon < 138 {
			return loadLocation("Australia/Adelaide") // ACST/ACDT (+09:30/+10:30)
		} else if lon < 147 {
			return loadLocation("Australia/Brisbane") // AEST (+10:00 no DST)
		} else {
			return loadLocation("Australia/Sydney") // AEST/AEDT (+10:00/+11:00)
		}
	}

	// New Zealand
	if lat >= -48 && lat <= -34 && lon >= 166 && lon <= 179 {
		return loadLocation("Pacific/Auckland") // NZST/NZDT (+12:00/+13:00)
	}

	// Japan
	if lat >= 24 && lat <= 46 && lon >= 123 && lon <= 146 {
		return loadLocation("Asia/Tokyo") // JST (+09:00)
	}

	// China
	if lat >= 18 && lat <= 54 && lon >= 73 && lon <= 135 {
		return loadLocation("Asia/Shanghai") // CST (+08:00)
	}

	// India
	if lat >= 8 && lat <= 37 && lon >= 68 && lon <= 97 {
		return loadLocation("Asia/Kolkata") // IST (+05:30)
	}

	// Middle East
	if lat >= 12 && lat <= 42 && lon >= 34 && lon <= 63 {
		return loadLocation("Asia/Dubai") // GST (+04:00)
	}

	// South America - Brazil
	if lat >= -34 && lat <= 5 && lon >= -74 && lon <= -35 {
		if lon > -50 {
			return loadLocation("America/Sao_Paulo") // BRT/BRST (-03:00/-02:00)
		} else {
			return loadLocation("America/Manaus") // AMT (-04:00)
		}
	}

	// South America - Argentina
	if lat >= -55 && lat <= -22 && lon >= -74 && lon <= -53 {
		return loadLocation("America/Argentina/Buenos_Aires") // ART (-03:00)
	}

	// Central America & Mexico
	if lat >= 14 && lat <= 33 && lon >= -118 && lon <= -86 {
		return loadLocation("America/Mexico_City") // CST/CDT (-06:00/-05:00)
	}

	// Default: Use UTC offset approximation based on longitude
	// Each 15° of longitude ≈ 1 hour of time difference
	// Note: Negative because Etc/GMT zones have reversed signs
	offsetName := fmt.Sprintf("Etc/GMT%+d", -hourOffset)
	return loadLocation(offsetName)
}

func loadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// Fallback to UTC if timezone not found
		return time.UTC
	}
	return loc
}
