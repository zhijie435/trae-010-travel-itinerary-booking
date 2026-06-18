package controllers

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func generateOrderNo() string {
	return fmt.Sprintf("ORD%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:4])
}

func generateRefundNo() string {
	return fmt.Sprintf("REF%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:4])
}

func timeNow() time.Time {
	return time.Now()
}
