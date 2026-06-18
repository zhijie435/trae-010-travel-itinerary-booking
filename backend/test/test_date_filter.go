package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Trip struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string
	Description string
	Destination string
	StartDate   time.Time
	EndDate     time.Time
	Price       float64
	TotalSpots  int
	LeftSpots   int
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func parseDate(dateStr string) (time.Time, error) {
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

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Trip{})

	testTrips := []Trip{
		{Name: "六月团-纯6月", StartDate: mustParse("2026-06-10"), EndDate: mustParse("2026-06-15")},
		{Name: "跨月团-六七月", StartDate: mustParse("2026-06-28"), EndDate: mustParse("2026-07-03")},
		{Name: "七月团-纯7月", StartDate: mustParse("2026-07-10"), EndDate: mustParse("2026-07-15")},
		{Name: "边界跨月团-6月30到7月1日", StartDate: mustParse("2026-06-30"), EndDate: mustParse("2026-07-01")},
		{Name: "跨月团-五六月份", StartDate: mustParse("2026-05-28"), EndDate: mustParse("2026-06-03")},
		{Name: "云南大理5日游", StartDate: mustParse("2026-07-18"), EndDate: mustParse("2026-07-22")},
	}

	for _, t := range testTrips {
		db.Create(&t)
	}

	fmt.Println("=== 数据库中的日期格式 ===")
	var allTrips []Trip
	db.Find(&allTrips)
	for _, t := range allTrips {
		fmt.Printf("  %s: start=%s, end=%s\n", t.Name, t.StartDate.Format(time.RFC3339), t.EndDate.Format(time.RFC3339))
	}

	fmt.Println("\n=== 测试1: 直接字符串比较（旧方式）===")
	testStringCompare(db, "2026-06-01", "2026-06-30")

	fmt.Println("\n=== 测试2: 使用 julianday 比较（新方式）===")
	testJulianDayCompare(db, "2026-06-01", "2026-06-30")

	fmt.Println("\n=== 测试3: 直接用 time.Time 参数（GORM方式）===")
	testGormTimeCompare(db, "2026-06-01", "2026-06-30")

	fmt.Println("\n=== 测试4: 查看实际存储的原始值 ===")
	rows, _ := db.Raw("SELECT id, name, start_date, end_date, typeof(start_date), typeof(end_date) FROM trips").Rows()
	for rows.Next() {
		var id uint
		var name, startDate, endDate, startType, endType string
		rows.Scan(&id, &name, &startDate, &endDate, &startType, &endType)
		fmt.Printf("  %s: start=[%s] (%s), end=[%s] (%s)\n", name, startDate, startType, endDate, endType)
	}
	rows.Close()

	fmt.Println("\n=== 测试5: julianday 值比较 ===")
	rows, _ = db.Raw(`
		SELECT name, 
		       julianday(start_date) as jd_start,
		       julianday(end_date) as jd_end,
		       julianday('2026-06-01 00:00:00+08:00') as jd_filter_start,
		       julianday('2026-06-30 23:59:59+08:00') as jd_filter_end
		FROM trips
	`).Rows()
	for rows.Next() {
		var name string
		var jdStart, jdEnd, jdFilterStart, jdFilterEnd float64
		rows.Scan(&name, &jdStart, &jdEnd, &jdFilterStart, &jdFilterEnd)
		overlap := jdEnd >= jdFilterStart && jdStart <= jdFilterEnd
		fmt.Printf("  %s: jd_start=%.5f, jd_end=%.5f, overlap=%v\n", name, jdStart, jdEnd, overlap)
	}
	rows.Close()
}

func mustParse(dateStr string) time.Time {
	t, err := parseDate(dateStr)
	if err != nil {
		panic(err)
	}
	return t
}

func testStringCompare(db *gorm.DB, startDate, endDate string) {
	tStart, _ := parseDate(startDate)
	tEnd, _ := parseDate(endDate)
	endOfDay := tEnd.Add(24*time.Hour - time.Second)

	var trips []Trip
	query := db
	query = query.Where("end_date >= ?", tStart.Format("2006-01-02 15:04:05-07:00"))
	query = query.Where("start_date <= ?", endOfDay.Format("2006-01-02 15:04:05-07:00"))
	query.Find(&trips)

	fmt.Printf("  筛选范围: %s ~ %s\n", startDate, endDate)
	fmt.Printf("  查询参数: end_date >= '%s', start_date <= '%s'\n",
		tStart.Format("2006-01-02 15:04:05-07:00"),
		endOfDay.Format("2006-01-02 15:04:05-07:00"))
	fmt.Printf("  结果数量: %d\n", len(trips))
	for _, t := range trips {
		fmt.Printf("    - %s (%s ~ %s)\n", t.Name, t.StartDate.Format("2006-01-02"), t.EndDate.Format("2006-01-02"))
	}

	hasCrossMonth := false
	for _, t := range trips {
		if t.Name == "跨月团-六七月" || t.Name == "边界跨月团-6月30到7月1日" || t.Name == "跨月团-五六月份" {
			hasCrossMonth = true
			break
		}
	}
	if hasCrossMonth {
		fmt.Println("  ✓ 跨月团被正确包含")
	} else {
		fmt.Println("  ✗ 错误：跨月团被遗漏了！")
	}
}

func testJulianDayCompare(db *gorm.DB, startDate, endDate string) {
	tStart, _ := parseDate(startDate)
	tEnd, _ := parseDate(endDate)
	endOfDay := tEnd.Add(24*time.Hour - time.Second)

	var trips []Trip
	query := db
	query = query.Where("julianday(end_date) >= julianday(?)", tStart.Format("2006-01-02 15:04:05-07:00"))
	query = query.Where("julianday(start_date) <= julianday(?)", endOfDay.Format("2006-01-02 15:04:05-07:00"))
	query.Find(&trips)

	fmt.Printf("  筛选范围: %s ~ %s\n", startDate, endDate)
	fmt.Printf("  结果数量: %d\n", len(trips))
	for _, t := range trips {
		fmt.Printf("    - %s (%s ~ %s)\n", t.Name, t.StartDate.Format("2006-01-02"), t.EndDate.Format("2006-01-02"))
	}

	hasCrossMonth := false
	for _, t := range trips {
		if t.Name == "跨月团-六七月" || t.Name == "边界跨月团-6月30到7月1日" || t.Name == "跨月团-五六月份" {
			hasCrossMonth = true
			break
		}
	}
	if hasCrossMonth {
		fmt.Println("  ✓ 跨月团被正确包含")
	} else {
		fmt.Println("  ✗ 错误：跨月团被遗漏了！")
	}
}

func testGormTimeCompare(db *gorm.DB, startDate, endDate string) {
	tStart, _ := parseDate(startDate)
	tEnd, _ := parseDate(endDate)
	endOfDay := tEnd.Add(24*time.Hour - time.Second)

	var trips []Trip
	query := db
	query = query.Where("end_date >= ?", tStart)
	query = query.Where("start_date <= ?", endOfDay)
	query.Find(&trips)

	fmt.Printf("  筛选范围: %s ~ %s\n", startDate, endDate)
	fmt.Printf("  结果数量: %d\n", len(trips))
	for _, t := range trips {
		fmt.Printf("    - %s (%s ~ %s)\n", t.Name, t.StartDate.Format("2006-01-02"), t.EndDate.Format("2006-01-02"))
	}

	hasCrossMonth := false
	for _, t := range trips {
		if t.Name == "跨月团-六七月" || t.Name == "边界跨月团-6月30到7月1日" || t.Name == "跨月团-五六月份" {
			hasCrossMonth = true
			break
		}
	}
	if hasCrossMonth {
		fmt.Println("  ✓ 跨月团被正确包含")
	} else {
		fmt.Println("  ✗ 错误：跨月团被遗漏了！")
	}
}
