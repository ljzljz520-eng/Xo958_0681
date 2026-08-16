package fixture

import (
	"handmade-soap-shop/internal/memory"
	"handmade-soap-shop/internal/shop"
)

func NewService() *shop.Service {
	catalog := memory.NewCatalog([]shop.Soap{
		{Slug: "forest-morning", Name: "林间清晨", Scent: "雪松、松针与薄荷", Story: "冷制皂在木模中静置四十八小时，再经过六周熟成，留下雨后森林般干净的气息。", PriceCents: 4800},
		{Slug: "citrus-tea", Name: "柑橘茶席", Scent: "甜橙、佛手柑与红茶", Story: "灵感来自午后茶席，以橄榄油和乳木果油打底，香气明亮而不过分甜腻。", PriceCents: 5200},
		{Slug: "lavender-night", Name: "薰衣草夜", Scent: "真正薰衣草与岩兰草", Story: "每一批都由小锅低温制作，深沉草木气息适合一天结束后的安静时刻。", PriceCents: 5600},
	})
	accounts := memory.NewAccounts([]shop.Account{
		{ID: "member-001", Name: "林青", Email: "member@example.com", PasswordHash: shop.HashPassword("soap1234")},
	})
	notes := memory.NewMemberNotes([]memory.MemberNoteSeed{
		{AccountID: "member-001", Title: "你的秋季用皂说明", Body: "建议从林间清晨开始，每次使用后保持皂体通风干燥。会员补货日为每月第一个周五。"},
	})
	return shop.NewService(catalog, accounts, notes, memory.NewSessions())
}
