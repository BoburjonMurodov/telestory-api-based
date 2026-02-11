package i18n

var Locales = map[string]map[string]string{
	"en": {
		"welcome":     "🇺🇸 Welcome! Please choose your language:",
		"registered":  "Language set to English 🇺🇸",
		"instruction": "**You can send:**\n- `username` or `@username`\n- `+1234567890`",
		"processing":  "⏳ Processing...",
		"error_limit": "🚫 Daily limit reached (3/3). Upgrade to Premium for unlimited searches!",
	},
	"uz": {
		"welcome":     "🇺🇿 Xush kelibsiz! Tilni tanlang:",
		"registered":  "O'zbek tili tanlandi 🇺🇿",
		"instruction": "**Yuborishingiz mumkin:**\n- `username` yoki `@username`\n- `+998901234567`",
		"processing":  "⏳ Qidirilmoqda...",
		"error_limit": "🚫 Limit tugadi (3/3). Cheksiz qidirish uchun Premium oling!",
	},
	"ru": {
		"welcome":     "🇷🇺 Добро пожаловать! Выберите язык:",
		"registered":  "Язык выбран: Русский 🇷🇺",
		"instruction": "**Вы можете отправить:**\n- `username` или `@username`\n- `+79001234567`",
		"processing":  "⏳ Обработка...",
		"error_limit": "🚫 Лимит исчерпан (3/3). Купите Premium для безлимитного поиска!",
	},
}

func GetMessage(lang, key string) string {
	if lang == "" {
		lang = "en"
	}
	if texts, ok := Locales[lang]; ok {
		if msg, ok := texts[key]; ok {
			return msg
		}
	}
	// Fallback to English
	if msg, ok := Locales["en"][key]; ok {
		return msg
	}
	return key
}

// Supported Languages
const (
	LangEN = "en"
	LangUZ = "uz"
	LangRU = "ru"
)
