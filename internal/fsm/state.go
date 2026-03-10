package fsm

const (
	StateNone = "" // нет активного диалога → главное меню

	// Регистрация нового пользователя
	StateRegAwaitName  = "reg:await_name"
	StateRegAwaitPhone = "reg:await_phone"
	StateRegAwaitClass = "reg:await_class"

	// Запись на занятие
	StateBookAwaitSubject = "book:await_subject"
	StateBookAwaitDate    = "book:await_date"
	StateBookAwaitSlot    = "book:await_slot"
	StateBookConfirm      = "book:confirm"

	// Отмена записи
	StateCancelAwaitBooking = "cancel:await_booking"

	// Перенос записи
	StateRescheduleAwaitDate = "reschedule:await_date"
	StateRescheduleAwaitSlot = "reschedule:await_slot"

	// Редактирование профиля
	StateProfileAwaitName  = "profile:await_name"
	StateProfileAwaitPhone = "profile:await_phone"
	StateProfileAwaitClass = "profile:await_class"

	// Обращение к администратору (для незарегистрированных пользователей)
	StateContactAwaitMessage = "contact:await_message"

	// Административное управление слотами
	StateAdminAwaitDate    = "admin:await_date"
	StateAdminAwaitTime    = "admin:await_time"
	StateAdminAwaitSubject = "admin:await_subject"
	StateAdminAwaitGroup   = "admin:await_group"
)
