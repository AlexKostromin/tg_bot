package bot

import (
	"github.com/AlexKostromin/tg_bot/internal/fsm"
	"github.com/AlexKostromin/tg_bot/internal/mq"
	"github.com/AlexKostromin/tg_bot/internal/repository"
	"github.com/AlexKostromin/tg_bot/internal/service"
)

type Dependencies struct {
	UserRepo    *repository.UserRepository
	SlotRepo    *repository.SlotRepository
	SubjectRepo *repository.SubjectRepository
	BookingRepo *repository.BookingRepository
	FSM         *fsm.Storage
	UserSvc     *service.UserService
	BookingSvc  *service.BookingService
	Publisher   *mq.Publisher
	AdminChatID int64
}
