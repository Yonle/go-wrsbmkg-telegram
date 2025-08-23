package main

import (
	"math"
)

func checkFilter(lat, lon, mag float64) (pass bool) {
	if len(config.RegionsFilter) == 0 {
		return true
	}

	for _, region := range config.RegionsFilter {
		dist := calcCrow(
			region.Coords.Latitude,
			region.Coords.Longitude,
			lat, lon,
		)

		if dist < region.MaxDistance {
			continue
		}

		if mag < region.MinMagnitude {
			continue
		}

		pass = true
		break
	}
	return
}

func calcCrow(lat1, lon1, lat2, lon2 float64) float64 {
	R := 6371000.0
	rad := math.Pi / 180.0
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	lat1 = lat1 * rad
	lat2 = lat2 * rad
	a := math.Pow(math.Sin(dLat/2), 2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLon/2), 2)
	return R * 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
