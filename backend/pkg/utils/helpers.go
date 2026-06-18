package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func GenerateOrderNo() string {
	return fmt.Sprintf("ORD%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:4])
}

func GenerateRefundNo() string {
	return fmt.Sprintf("REF%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:4])
}

func TimeNow() time.Time {
	return time.Now()
}

func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}
