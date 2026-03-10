package keyboard

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// MainMenu возвращает reply-клавиатуру главного меню.
func MainMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Записаться"),
			tgbotapi.NewKeyboardButton("📋 Мои записи"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👤 Профиль"),
		),
	)
}

// ClassNumbers возвращает клавиатуру с номерами классов 5–11.
func ClassNumbers() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("5"),
			tgbotapi.NewKeyboardButton("6"),
			tgbotapi.NewKeyboardButton("7"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("8"),
			tgbotapi.NewKeyboardButton("9"),
			tgbotapi.NewKeyboardButton("10"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("11"),
		),
	)
	kb.OneTimeKeyboard = true
	return kb
}
