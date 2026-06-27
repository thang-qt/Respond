package i18n

var tagNamesVI = map[string]string{
	"philosophy":    "Triết học",
	"ethics":        "Đạo đức",
	"politics":      "Chính trị",
	"law":           "Luật pháp",
	"economics":     "Kinh tế",
	"business":      "Kinh doanh",
	"technology":    "Công nghệ",
	"ai":            "AI",
	"science":       "Khoa học",
	"health":        "Sức khỏe",
	"environment":   "Môi trường",
	"education":     "Giáo dục",
	"psychology":    "Tâm lý học",
	"society":       "Xã hội",
	"culture":       "Văn hóa",
	"art":           "Nghệ thuật",
	"history":       "Lịch sử",
	"religion":      "Tôn giáo",
	"sports":        "Thể thao",
	"international": "Quốc tế",
	"security":      "An ninh",
	"lifestyle":     "Lối sống",
	"future":        "Tương lai",
	"meta":          "Meta",
}

func LocalizeTagName(locale, slug, fallback string) string {
	if NormalizeLocale(locale) == LocaleVI {
		if name, ok := tagNamesVI[slug]; ok {
			return name
		}
	}
	return fallback
}
