package controllers

import (
	"net/http"
	"travel-refund/config"
	"travel-refund/models"
	"travel-refund/pkg/utils"

	"github.com/gin-gonic/gin"
)

type SeedController struct{}

func NewSeedController() *SeedController {
	return &SeedController{}
}

func (ctrl *SeedController) SeedData(c *gin.Context) {
	users := []models.User{
		{Username: "admin", Email: "admin@test.com", Phone: "13800000000", Role: "admin"},
		{Username: "user1", Email: "user1@test.com", Phone: "13800000001", Role: "user"},
		{Username: "user2", Email: "user2@test.com", Phone: "13800000002", Role: "user"},
	}
	for _, u := range users {
		config.DB.FirstOrCreate(&u, models.User{Username: u.Username})
	}

	now := utils.TimeNow()
	routes := []models.Route{
		{Name: "云南大理5日游", Description: "探索风花雪月的大理古城", Destination: "云南大理",
			StartDate: now.AddDate(0, 1, 0), EndDate: now.AddDate(0, 1, 4), Price: 2999, Status: "active"},
		{Name: "三亚海岛度假游", Description: "阳光沙滩，椰林海风", Destination: "海南三亚",
			StartDate: now.AddDate(0, 2, 0), EndDate: now.AddDate(0, 2, 6), Price: 4599, Status: "active"},
		{Name: "西藏拉萨朝圣之旅", Description: "雪域高原，心灵之旅", Destination: "西藏拉萨",
			StartDate: now.AddDate(0, 3, 0), EndDate: now.AddDate(0, 3, 7), Price: 6899, Status: "active"},
	}

	tx := config.DB.Begin()
	for _, r := range routes {
		var existing models.Route
		result := tx.Where("name = ?", r.Name).First(&existing)
		if result.Error != nil {
			if err := tx.Create(&r).Error; err == nil {
				totalSpots := 20
				if r.Name == "三亚海岛度假游" {
					totalSpots = 30
				} else if r.Name == "西藏拉萨朝圣之旅" {
					totalSpots = 15
				}
				inventory := &models.Inventory{
					RouteID:    r.ID,
					Date:       r.StartDate.Format("2006-01-02"),
					TotalSpots: totalSpots,
					LeftSpots:  totalSpots - 5,
					Status:     "active",
				}
				tx.Create(inventory)

				adjustLog := &models.InventoryAdjustLog{
					InventoryID:  inventory.ID,
					RouteID:      r.ID,
					AdjustType:   "init",
					OldSpots:     0,
					NewSpots:     inventory.LeftSpots,
					AdjustAmount: totalSpots,
					Reason:       "线路创建初始化",
				}
				tx.Create(adjustLog)
			}
		}
	}
	tx.Commit()

	var route1, route2, route3 models.Route
	config.DB.Where("name = ?", "云南大理5日游").First(&route1)
	config.DB.Where("name = ?", "三亚海岛度假游").First(&route2)
	config.DB.Where("name = ?", "西藏拉萨朝圣之旅").First(&route3)

	daliItineraries := []models.Itinerary{
		{RouteID: route1.ID, DayNumber: 1, Title: "出发抵达大理", Breakfast: "自理", Lunch: "自理", Dinner: "大理特色白族菜",
			Accommodation: "大理古城特色客栈", Transportation: "飞机+专车接送",
			Activities: "抵达大理机场，专车接往古城，自由活动逛古城夜景", Notes: "建议携带防晒用品"},
		{RouteID: route1.ID, DayNumber: 2, Title: "洱海环游", Breakfast: "酒店早餐", Lunch: "洱海渔村特色菜", Dinner: "双廊古镇美食",
			Accommodation: "双廊海景酒店", Transportation: "商务车环湖",
			Activities: "洱海骑行、喜洲古镇、双廊艺术小镇", Notes: "注意防晒，做好高原反应预防"},
		{RouteID: route1.ID, DayNumber: 3, Title: "苍山索道+古城深度游", Breakfast: "酒店早餐", Lunch: "大理风味餐厅", Dinner: "特色过桥米线",
			Accommodation: "大理古城特色客栈", Transportation: "商务车+索道",
			Activities: "苍山洗马潭索道、天龙八部影视城、洋人街", Notes: "山顶温度较低，建议携带外套"},
		{RouteID: route1.ID, DayNumber: 4, Title: "崇圣寺三塔+返程", Breakfast: "酒店早餐", Lunch: "景区餐厅", Dinner: "自理",
			Accommodation: "无", Transportation: "商务车+飞机",
			Activities: "崇圣寺三塔文化旅游区，送机返程", Notes: "退房时间中午12点前"},
	}
	for _, it := range daliItineraries {
		config.DB.FirstOrCreate(&it, models.Itinerary{RouteID: it.RouteID, DayNumber: it.DayNumber})
	}

	sanyaItineraries := []models.Itinerary{
		{RouteID: route2.ID, DayNumber: 1, Title: "抵达三亚，入住酒店", Breakfast: "自理", Lunch: "自理", Dinner: "海鲜大餐",
			Accommodation: "三亚湾海景酒店", Transportation: "飞机+专车接送",
			Activities: "抵达三亚凤凰机场，入住酒店，自由活动", Notes: "做好防晒措施"},
		{RouteID: route2.ID, DayNumber: 2, Title: "蜈支洲岛一日游", Breakfast: "酒店早餐", Lunch: "岛上海鲜餐厅", Dinner: "沙滩BBQ",
			Accommodation: "三亚湾海景酒店", Transportation: "游船+专车",
			Activities: "蜈支洲岛观光、海上项目可选（潜水、摩托艇等）", Notes: "海上项目需自费"},
		{RouteID: route2.ID, DayNumber: 3, Title: "热带雨林+南山寺", Breakfast: "酒店早餐", Lunch: "素斋", Dinner: "椰林特色餐",
			Accommodation: "三亚湾海景酒店", Transportation: "商务车",
			Activities: "呀诺达热带雨林、南山文化旅游区、108米海上观音", Notes: "建议穿舒适运动鞋"},
		{RouteID: route2.ID, DayNumber: 4, Title: "免税购物+返程", Breakfast: "酒店早餐", Lunch: "特色小吃", Dinner: "自理",
			Accommodation: "无", Transportation: "专车+飞机",
			Activities: "三亚国际免税城购物，送机返程", Notes: "记得带身份证购物"},
	}
	for _, it := range sanyaItineraries {
		config.DB.FirstOrCreate(&it, models.Itinerary{RouteID: it.RouteID, DayNumber: it.DayNumber})
	}

	lhasaItineraries := []models.Itinerary{
		{RouteID: route3.ID, DayNumber: 1, Title: "抵达拉萨，适应高原", Breakfast: "自理", Lunch: "藏式特色餐", Dinner: "酒店餐厅",
			Accommodation: "拉萨市区供氧酒店", Transportation: "飞机+专车接送",
			Activities: "抵达拉萨，酒店休息适应高原环境", Notes: "不要剧烈运动，多喝温水"},
		{RouteID: route3.ID, DayNumber: 2, Title: "布达拉宫+大昭寺", Breakfast: "酒店早餐", Lunch: "藏式餐厅", Dinner: "特色小吃",
			Accommodation: "拉萨市区供氧酒店", Transportation: "商务车",
			Activities: "布达拉宫参观、大昭寺转经、八廓街逛", Notes: "参观布达拉宫需提前预约"},
		{RouteID: route3.ID, DayNumber: 3, Title: "羊卓雍措一日游", Breakfast: "酒店早餐", Lunch: "景区餐厅", Dinner: "酒店餐厅",
			Accommodation: "拉萨市区供氧酒店", Transportation: "商务车",
			Activities: "圣湖羊卓雍措、卡若拉冰川", Notes: "海拔较高，注意保暖和高反"},
		{RouteID: route3.ID, DayNumber: 4, Title: "纳木措一日游", Breakfast: "酒店早餐", Lunch: "简餐", Dinner: "特色餐",
			Accommodation: "拉萨市区供氧酒店", Transportation: "商务车",
			Activities: "天湖纳木措、那根拉山口", Notes: "路程较远，建议备晕车药"},
		{RouteID: route3.ID, DayNumber: 5, Title: "返程", Breakfast: "酒店早餐", Lunch: "自理", Dinner: "自理",
			Accommodation: "无", Transportation: "专车+飞机",
			Activities: "自由活动，送机返程", Notes: "检查好随身物品"},
	}
	for _, it := range lhasaItineraries {
		config.DB.FirstOrCreate(&it, models.Itinerary{RouteID: it.RouteID, DayNumber: it.DayNumber})
	}

	c.JSON(http.StatusOK, gin.H{"message": "测试数据初始化成功"})
}
