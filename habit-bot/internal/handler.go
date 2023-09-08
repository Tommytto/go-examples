package internal

import (
	"errors"
	"fmt"
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/Tommytto/habit-bot/internal/helpers"
	"github.com/Tommytto/habit-bot/internal/repos"
	"github.com/Tommytto/habit-bot/internal/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BotApi interface {
}

type Handler struct {
	bot               BotAPI
	userService       *services.UserService
	habitService      *services.HabitService
	motivationService *services.MotivationService
}

func NewHandler(bot BotAPI, userService *services.UserService, habitService *services.HabitService, motivationService *services.MotivationService) *Handler {
	return &Handler{bot: bot, userService: userService, habitService: habitService, motivationService: motivationService}
}

func (h *Handler) HandleUpdate(update *tgbotapi.Update) error {
	chatId, err := h.GetChatId(update)
	if err != nil {
		log.Print("chatId not found", err)
		return err
	}
	telegramUser, err := h.ExtractUser(update)
	if err != nil {
		log.Print("userId not found", err)
		h.SendText(chatId, "Не указан telegramId")
		return err
	}

	user, err := h.userService.FindOrCreateByTelegramId(&services.CreateUserInput{
		TelegramId:       telegramUser.ID,
		ChatId:           chatId,
		FirstName:        telegramUser.FirstName,
		LastName:         telegramUser.LastName,
		TelegramUsername: telegramUser.UserName,
	})
	if err != nil {
		log.Print(err)
		h.EMessage(chatId)
		return err
	}

	if update.Message != nil {
		err := h.HandleTextMessage(update, user)
		if err != nil {
			log.Print(err)
			return err
		}
	} else if update.CallbackQuery != nil {
		err := h.HandleCallback(update, user)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}

func (h *Handler) GetChatId(update *tgbotapi.Update) (int64, error) {
	if update.Message != nil {
		return update.Message.Chat.ID, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID, nil
	}

	return 0, fmt.Errorf("chat id not found")
}

func (h *Handler) ExtractUser(update *tgbotapi.Update) (*tgbotapi.User, error) {
	if update.Message != nil {
		return update.Message.From, nil
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.From, nil
	}

	return nil, fmt.Errorf("user id not found")
}

const (
	StepCreateHabit  = "step_create_habit"
	StepWriteThought = "step_write_thought"
)

const (
	KeyboardAffirmation = "affirmation"
	KeyboardComplete    = "complete"
	KeyboardArchive     = "archive"
)

func (h *Handler) HandleCallback(update *tgbotapi.Update, user *repos.UserEntity) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := h.bot.Request(callback); err != nil {
		h.EMessage(chatId)
		log.Print(err)
		return err
	}

	dataFields := strings.Fields(update.CallbackQuery.Data)
	if len(dataFields) == 0 {
		return fmt.Errorf("bad data query provided %s", update.CallbackQuery.Data)
	}
	command := dataFields[0]
	args := strings.Join(dataFields[1:], " ")
	switch command {
	case KeyboardComplete:
		habitId := args
		err := h.CommandCompleteHabit(chatId, habitId, user, update.CallbackQuery.Message.Time())
		if err != nil {
			return err
		}
		if h.habitService.NeedAskThought(habitId, update.CallbackQuery.Message.Time()) {
			err := h.CommandAskThought(habitId, user)
			if err != nil {
				return err
			}
			return nil
		} else {
			err = h.SendAffirmationIfHaveNot(user, update.CallbackQuery.Message.Time())
			if err != nil {
				return err
			}
		}
	case KeyboardAffirmation:
		if h.userService.GetAndSetAffirmationToday(user, update.CallbackQuery.Message.Time()) {
			h.SendText(chatId, "Сегодня уже получал, приходи завтра :)")
			return nil
		}
		return h.SendText(chatId, h.motivationService.RandomAffirmation())
	case KeyboardArchive:
		habitId := args
		err := h.habitService.ToggleArchive(habitId)
		if err != nil {
			return err
		}

		keyboard, err := h.GetArchiveKeyboard(user)
		if err != nil {
			return err
		}

		editedMsg := tgbotapi.NewEditMessageReplyMarkup(chatId, update.CallbackQuery.Message.MessageID, *keyboard)

		h.SendMessage(editedMsg)
	}

	return nil
}

func (h *Handler) HandleTextMessage(update *tgbotapi.Update, user *repos.UserEntity) error {
	msg := update.Message
	chatId := msg.Chat.ID

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			h.SendText(chatId, fmt.Sprintf("Добро пожаловать в трекер, %s\n\n", msg.From.FirstName))
			h.SendText(chatId, fmt.Sprintf("В списке команд можно увидеть команды увидеть привычки, добавить их и трекать!"))
			h.SendText(chatId, fmt.Sprintf("Можешь начать с /create_habit"))
			return nil
		case "id":
			return h.SendText(chatId, fmt.Sprintf("id: %s\ntelegramId: %v", user.Id, user.TelegramId))
		case "ok":
			newMsg := tgbotapi.NewMessage(chatId, "ok")
			newMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
			return h.SendMessage(newMsg)
		case "status":
			return h.CommandState(chatId, user.Id, msg.Time())
		case "create_habit":
			return h.CommandIntroCreateHabit(chatId, user)
		case "archive":
			return h.CommandShowArchive(chatId, user)
		case "help":
			err := h.CommandHelp(chatId)
			if err != nil {
				return err
			}
			return nil
		default:
			return h.SendText(chatId, "Try /help")
		}
	}
	if user.Step == StepCreateHabit {
		err := h.CommandCreateHabit(msg, chatId, user)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(user.Step, StepWriteThought) {
		if habitId, err := extractHabitIdFromStep(user.Step); err != nil {
			if err := h.userService.ResetStep(user); err != nil {
				h.EMessage(chatId)
				log.Print("can't reset user step", err)
				return err
			}
			log.Print("step reseted for ", user.Id)
			return err
		} else {
			err := h.CommandWriteThoughtHandle(msg, habitId, user)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return h.CommandSendKeyboard(chatId, user)
}

func (h *Handler) CommandSendKeyboard(chatId int64, user *repos.UserEntity) error {
	inputMessage := tgbotapi.NewMessage(chatId, "Привет ❤️ Выбирай что выполнил")

	err := h.ApplyHabitsKeyboard(&inputMessage, user.Id)
	if err != nil {
		fmt.Println("err", err)
		return err
	}

	return h.SendMessage(inputMessage)
}

func (h *Handler) ApplyHabitsKeyboard(msg *tgbotapi.MessageConfig, id repos.UserId) error {
	habits, err := h.habitService.GetAllActive(id)
	if err != nil {
		h.EMessage(msg.ChatID)
		return err
	}

	// add buttons for completing habits
	if len(habits) > 0 {
		var keyboardRows [][]tgbotapi.InlineKeyboardButton
		for _, habit := range habits {
			keyboardRows = append(
				keyboardRows,
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						fmt.Sprintf("Подтвердить %s", habit.Name),
						KeyboardComplete+" "+habit.Id,
					)))
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

		msg.ReplyMarkup = keyboard
	}

	return nil
}

func (h *Handler) CommandCompleteHabitCongratulation(chatId int64, habitId string, when time.Time) error {
	statusText, err := h.GetHabitStatusText(habitId, when)
	if err != nil {
		fmt.Println("GetHabitStatusText problem", habitId)
		h.EMessage(chatId)
		return err
	}
	text := ""
	text += h.motivationService.RandomKudo()
	text += "\n\n"
	text += statusText

	h.SendText(chatId, text)

	return nil
}

func (h *Handler) SendAffirmationIfHaveNot(user *repos.UserEntity, when time.Time) error {
	if !h.userService.AffirmationWasSentToday(user, when) {
		time.Sleep(300)
		msg := tgbotapi.NewMessage(user.ChatId, "Выбери случайную аффирмацию на день за то что ты молодец!")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("1"), KeyboardAffirmation+" 1",
				),
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("2"), KeyboardAffirmation+" 2",
				),
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("3"), KeyboardAffirmation+" 3",
				),
			),
		)
		return h.SendMessage(msg)
	}
	return nil
}

func (h *Handler) GetHabitStatusText(habitId string, when time.Time) (string, error) {
	streak, err := h.habitService.GetCurrentStreakLength(habitId, when)
	if err != nil {
		fmt.Println("can't get current streak", err)
		return "", err
	}

	habit, err := h.habitService.GetOne(habitId)
	if err != nil {
		return "", err
	}
	dayWord := helpers.Declension(streak, []string{"день", "дня", "дней"})
	text := ""
	text += fmt.Sprintf("%s - %v %v без перерыва!\n", habit.Name, streak, dayWord)
	return text, nil
}

func (h *Handler) CommandState(chatId int64, userId string, when time.Time) error {
	habits, err := h.habitService.GetAllActive(userId)
	if err != nil {
		fmt.Println("can't get habits", err)
		return err
	}

	type wrapper struct {
		StreakLength int
		Name         string
	}
	var habitsData []*wrapper
	for _, habit := range habits {
		result, err := h.habitService.GetCurrentStreakLength(habit.Id, when)
		if err != nil {
			fmt.Println("can't get checkins", err)
			return err
		}
		habitsData = append(habitsData, &wrapper{
			result,
			habit.Name,
		})
	}

	text := ""
	text += "Твой статус\n"
	for _, habitsInfo := range habitsData {
		text += fmt.Sprintf("%s - %v без перерыва!\n", habitsInfo.Name, habitsInfo.StreakLength)
	}
	return h.SendText(chatId, text)
}

func (h *Handler) SendText(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)

	return h.SendMessage(msg)
}

func (h *Handler) SendMD(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"

	return h.SendMessage(msg)
}

func (h *Handler) SendMessage(msg tgbotapi.Chattable) error {
	if _, err := h.bot.Send(msg); err != nil {
		log.Panic(err)
		return err
	}

	return nil
}

func (h *Handler) EMessage(chatId int64) error {
	return h.SendText(chatId, "У бота проблемы")
}

func (h *Handler) CommandCreateHabit(msg *tgbotapi.Message, chatId int64, user *repos.UserEntity) error {
	r, _ := regexp.Compile(`(.*)\s+(\d+)`)
	createHabitTextParts := r.FindStringSubmatch(msg.Text)
	if len(createHabitTextParts) != 3 {
		h.SendText(chatId, "Введите название и количество дней со старта привычки (не считая сегодня, отметите сами)\n/create_habit нет сахару 50")
		err := fmt.Errorf("not enough words")
		log.Print(err)
		return err
	}
	habitName := createHabitTextParts[1]
	daysCompleted, err := strconv.Atoi(createHabitTextParts[2])
	if err != nil {
		h.SendText(chatId, "Введите название и количество дней со старта привычки\n нет сахару 50")
		err := fmt.Errorf("bad days completed token")
		log.Print(err)
		return err
	}
	_, err = h.habitService.CreateHabit(&services.CreateHabitInput{
		UserId:        user.Id,
		Name:          habitName,
		DaysCompleted: daysCompleted,
		CreatedAt:     msg.Time(),
	})
	if err != nil {
		if errors.Is(err, ers.ErrBadInput) {
			h.SendText(chatId, "Введите корректное название")
		} else {
			h.EMessage(chatId)
		}

		return err
	}
	if err := h.userService.ResetStep(user); err != nil {
		h.EMessage(chatId)
		return err
	}

	return h.SendText(chatId, "Успешно!")
}

func (h *Handler) CommandIntroCreateHabit(chatId int64, user *repos.UserEntity) error {
	err := h.userService.SetStep(user, StepCreateHabit)
	if err != nil {
		h.EMessage(chatId)
		return err
	}
	text := ""
	text += "Введите название привычки и в конце укажите сколько дней уже следуете ей (не считая сегодня)\n"
	text += "пример: \n"
	text += "без сахара 23"
	return h.SendText(chatId, text)
}

func (h *Handler) CommandCompleteHabit(chatId int64, habitId string, user *repos.UserEntity, completedAt time.Time) error {
	_, err := h.habitService.CompleteToday(habitId, completedAt)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyCompleted) {
			h.SendText(chatId, "Ты сегодня уже отмечал 💋")
			statusText, err := h.GetHabitStatusText(habitId, completedAt)
			if err != nil {
				log.Print(err)
				h.EMessage(chatId)
				return err
			}
			h.SendText(chatId, statusText)
			return nil
		} else {
			h.EMessage(chatId)
			return err
		}
	}

	return h.CommandCompleteHabitCongratulation(chatId, habitId, completedAt)
}

func (h *Handler) CommandHelp(chatId int64) error {
	commandsConfig := getCommandsConfig()

	text := ""
	for _, c := range commandsConfig.Commands {
		text += fmt.Sprintf("%s - %s\n", c.Command, c.Description)
	}
	text += "\n"
	text += "В случае проблем и пожеланий писать @tommytoo"
	return h.SendText(chatId, text)
}

func generateStepWriteThought(habitId string) string {
	return StepWriteThought + " " + habitId
}

func extractHabitIdFromStep(step string) (string, error) {
	fields := strings.Fields(step)
	if len(fields) < 2 {
		log.Print("bad step")
		return "", fmt.Errorf("bad step")
	}
	return fields[1], nil
}

func (h *Handler) CommandAskThought(habitId string, user *repos.UserEntity) error {
	if err := h.userService.SetStep(user, generateStepWriteThought(habitId)); err != nil {
		log.Print("can't save step write thought", err)
		return err
	}
	text := ""
	text += "Напиши свои ощущения и мысли\n"
	text += "Тяжело ли тебе сейчас или просто, в общем всё что есть в голове! 🥰"
	if err := h.SendText(user.ChatId, text); err != nil {
		log.Print("can't send text: ", err)
		return err
	}
	return nil
}

func (h *Handler) CommandWriteThoughtHandle(msg *tgbotapi.Message, habitId string, user *repos.UserEntity) error {
	if err := h.habitService.AddThought(habitId, msg.Text); err != nil {
		log.Print("can't add thought: ", err)
		return err
	}
	if err := h.userService.ResetStep(user); err != nil {
		log.Print("can't reset step", err)
		return err
	}

	h.SendText(user.ChatId, "Отлично, хорошего тебе дня!")
	return h.SendAffirmationIfHaveNot(user, msg.Time())
}

func (h *Handler) CommandShowArchive(chatId int64, user *repos.UserEntity) error {
	msg := tgbotapi.NewMessage(chatId, "")

	content := ""
	content += "**Активные привычки**"

	keyboard, err := h.GetArchiveKeyboard(user)
	if err != nil {
		return err
	}

	msg.Text = content
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"

	return h.SendMessage(msg)
}

func (h *Handler) GetArchiveKeyboard(user *repos.UserEntity) (*tgbotapi.InlineKeyboardMarkup, error) {
	habits, err := h.habitService.GetAll(user.Id)
	if err != nil {
		return nil, err
	}

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	for _, habit := range habits {
		archiveStatusIcon := "✅"
		if habit.Archived {
			archiveStatusIcon = ""
		}
		keyboardRows = append(
			keyboardRows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s %s", archiveStatusIcon, habit.Name),
					KeyboardArchive+" "+habit.Id,
				)))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	return &keyboard, nil
}

var CommandConfig = tgbotapi.NewSetMyCommands(
	tgbotapi.BotCommand{
		Command:     "/help",
		Description: "Помощь",
	},
	tgbotapi.BotCommand{
		Command:     "/create_habit",
		Description: "Добавить привычку",
	},
	tgbotapi.BotCommand{
		Command:     "/status",
		Description: "Статус по всем привычкам",
	},
	tgbotapi.BotCommand{
		Command:     "/archive",
		Description: "Спрятать старые привычки",
	},
)

func getCommandsConfig() tgbotapi.SetMyCommandsConfig {
	return CommandConfig
}

func SetupCommands(bot *tgbotapi.BotAPI) {
	_, _ = bot.Request(getCommandsConfig())
}

//go:generate mockgen -destination=mocks/mock_bot_api.go -package=mocks . BotAPI
type BotAPI interface {
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}
