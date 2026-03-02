package i18n

var Locales = map[string]map[string]string{
	"en": {
		"welcome":        "🇺🇸 Welcome! Please choose your language:",
		"registered":     "Language set to English 🇺🇸",
		"instruction":    "**You can send:**\n- `username` or `@username`\n- `+1234567890`",
		"processing":     "⏳ Processing...",
		"error_limit":    "🚫 Daily limit reached (%d/%d). Upgrade to Premium for unlimited searches!",
		"story_count":    "📊 Found %d stories for `%s`",
		"no_stories":     "📭 No stories found for `%s`",
		"fetch_error":    "❌ Error fetching stories: %s",
		"download_error": "⚠️ Some stories couldn't be downloaded. Sent %d of %d stories.",
		"downloading":    "📊 Found %d stories. Downloading...",
		"story_from":     "Story from %s",
		"cooldown":       "Please wait %d seconds between downloads.",
		"invalid_input":  "❌ Invalid input! Please send only a valid username (e.g. `@username`) or a phone number (e.g. `+1234567890`).",
		"stats_report": "📊 **Bot Analytics**\n\n" +
			"👥 **Total Users:** %d\n" +
			"🔥 **Active Users (7 Days):** %d\n" +
			"🔥 **Active Users (30 Days):** %d\n\n" +
			"📅 **Today's Downloads:** %d\n" +
			"✅ Success: %d | ❌ Failed: %d\n\n" +
			"📥 **Total Downloads (All-Time):** %d\n" +
			"✅ Success: %d | ❌ Failed: %d",
	},
	"uz": {
		"welcome":        "🇺🇿 Xush kelibsiz! Tilni tanlang:",
		"registered":     "O'zbek tili tanlandi 🇺🇿",
		"instruction":    "**Yuborishingiz mumkin:**\n- `username` yoki `@username`\n- `+998901234567`",
		"processing":     "⏳ Qidirilmoqda...",
		"error_limit":    "🚫 Limit tugadi (%d/%d). Cheksiz qidirish uchun Premium oling!",
		"story_count":    "📊 %d ta hikoya topildi — `%s`",
		"no_stories":     "📭 `%s` uchun hikoya topilmadi",
		"fetch_error":    "❌ Hikoyalarni yuklashda xatolik: %s",
		"download_error": "⚠️ Ba'zi hikoyalar yuklanmadi. %d/%d ta hikoya yuborildi.",
		"downloading":    "📊 %d ta hikoya topildi. Yuklanmoqda...",
		"story_from":     "%s dan hikoya",
		"cooldown":       "Iltimos, yuklashlar orasida %d soniya kuting.",
		"invalid_input":  "❌ Noto'g'ri format! Iltimos, faqat username (masalan, `@username`) yoki telefon raqami (masalan, `+998901234567`) yuboring.",
		"stats_report": "📊 **Bot Statistikasi**\n\n" +
			"👥 **Jami Foydalanuvchilar:** %d\n" +
			"🔥 **Faol Foydalanuvchilar (7 kun):** %d\n" +
			"🔥 **Faol Foydalanuvchilar (30 kun):** %d\n\n" +
			"📅 **Bugungi Yuklashlar:** %d\n" +
			"✅ Muvaffaqiyatli: %d | ❌ Xatoliklar: %d\n\n" +
			"📥 **Jami Yuklashlar (Barcha vaqt):** %d\n" +
			"✅ Muvaffaqiyatli: %d | ❌ Xatoliklar: %d",
	},
	"ru": {
		"welcome":        "🇷🇺 Добро пожаловать! Выберите язык:",
		"registered":     "Язык выбран: Русский 🇷🇺",
		"instruction":    "**Вы можете отправить:**\n- `username` или `@username`\n- `+79001234567`",
		"processing":     "⏳ Обработка...",
		"error_limit":    "🚫 Лимит исчерпан (%d/%d). Купите Premium для безлимитного поиска!",
		"story_count":    "📊 Найдено %d историй для `%s`",
		"no_stories":     "📭 Истории не найдены для `%s`",
		"fetch_error":    "❌ Ошибка загрузки историй: %s",
		"download_error": "⚠️ Некоторые истории не удалось загрузить. Отправлено %d из %d историй.",
		"downloading":    "📊 Найдено %d историй. Загрузка...",
		"story_from":     "История от %s",
		"cooldown":       "Пожалуйста, подождите %d секунд между загрузками.",
		"invalid_input":  "❌ Неверный ввод! Пожалуйста, отправьте только имя пользователя (например, `@username`) или номер телефона (например, `+79001234567`).",
		"stats_report": "📊 **Аналитика Бота**\n\n" +
			"👥 **Всего Пользователей:** %d\n" +
			"🔥 **Активные Пользователи (7 Дней):** %d\n" +
			"🔥 **Активные Пользователи (30 Дней):** %d\n\n" +
			"📅 **Загрузки за Сегодня:** %d\n" +
			"✅ Успешно: %d | ❌ Ошибки: %d\n\n" +
			"📥 **Всего Загрузок (За всё время):** %d\n" +
			"✅ Успешно: %d | ❌ Ошибки: %d",
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
