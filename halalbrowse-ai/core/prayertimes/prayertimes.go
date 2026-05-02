package prayertimes

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Schedule struct {
	Date     string            `json:"date"`
	Timezone string            `json:"timezone"`
	Prayers  map[string]string `json:"prayers"`
}

type Manager struct {
	PreWindowMinutes  int
	PostWindowMinutes int
	StrictDelta       float64
}

func (m Manager) EffectiveThreshold(base float64, strict bool) float64 {
	if !strict {
		return base
	}
	result := base - m.StrictDelta
	if result < 0.05 {
		return 0.05
	}
	return math.Round(result*100) / 100
}

func (m Manager) StrictModeAt(schedule Schedule, now time.Time) (bool, string) {
	pre := durationOrDefault(m.PreWindowMinutes, 15)
	post := durationOrDefault(m.PostWindowMinutes, 10)
	for prayer, value := range schedule.Prayers {
		prayerTime, err := parsePrayerTime(schedule, value)
		if err != nil {
			continue
		}
		if now.After(prayerTime.Add(-pre)) && now.Before(prayerTime.Add(post)) || now.Equal(prayerTime) {
			return true, prayer
		}
	}
	return false, ""
}

func FetchDaily(client *http.Client, city, country string, day time.Time) (Schedule, error) {
	if client == nil {
		client = http.DefaultClient
	}
	url := fmt.Sprintf("https://api.aladhan.com/v1/timingsByCity/%d-%02d-%02d?city=%s&country=%s&method=2", day.Year(), day.Month(), day.Day(), city, country)
	resp, err := client.Get(url)
	if err != nil {
		return Schedule{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Schedule{}, fmt.Errorf("aladhan returned status %d", resp.StatusCode)
	}
	var payload struct {
		Data struct {
			Timings map[string]string `json:"timings"`
			Meta    struct {
				Timezone string `json:"timezone"`
			} `json:"meta"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Schedule{}, err
	}
	prayers := map[string]string{}
	for _, name := range []string{"Fajr", "Dhuhr", "Asr", "Maghrib", "Isha"} {
		prayers[name] = cleanClock(payload.Data.Timings[name])
	}
	return Schedule{Date: day.Format("2006-01-02"), Timezone: payload.Data.Meta.Timezone, Prayers: prayers}, nil
}

func LoadFallbackCSV(content string, day time.Time, city string) (Schedule, error) {
	records, err := csv.NewReader(strings.NewReader(content)).ReadAll()
	if err != nil {
		return Schedule{}, err
	}
	for idx, row := range records {
		if idx == 0 || len(row) < 8 {
			continue
		}
		if row[0] == day.Format("2006-01-02") && strings.EqualFold(row[1], city) {
			return Schedule{
				Date:     row[0],
				Timezone: row[2],
				Prayers:  map[string]string{"Fajr": row[3], "Dhuhr": row[4], "Asr": row[5], "Maghrib": row[6], "Isha": row[7]},
			}, nil
		}
	}
	return Schedule{}, fmt.Errorf("no fallback prayer schedule for %s on %s", city, day.Format("2006-01-02"))
}

func parsePrayerTime(schedule Schedule, hhmm string) (time.Time, error) {
	parts := strings.Split(cleanClock(hhmm), ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid prayer time %q", hhmm)
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}
	location := time.UTC
	if schedule.Timezone != "" {
		if tz, err := time.LoadLocation(schedule.Timezone); err == nil {
			location = tz
		}
	}
	day, err := time.ParseInLocation("2006-01-02", schedule.Date, location)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, location), nil
}

func cleanClock(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.Index(value, " "); idx > 0 {
		value = value[:idx]
	}
	return value
}

func durationOrDefault(minutes int, fallback int) time.Duration {
	if minutes <= 0 {
		minutes = fallback
	}
	return time.Duration(minutes) * time.Minute
}
