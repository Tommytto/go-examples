package main

import (
	"fmt"
	"github.com/Tommytto/habit-bot/internal"
	"github.com/Tommytto/habit-bot/internal/aws"
	"github.com/Tommytto/habit-bot/internal/common"
	"github.com/Tommytto/habit-bot/internal/repos"
	"github.com/Tommytto/habit-bot/internal/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"time"
)

var errors = 0
var sent = 0
var skipped = 0

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	awsWrapper := aws.NewAwsWrapper()
	db := awsWrapper.GetDynamo()
	usersRepo := repos.NewUsersRepoDynamo(db)
	habitsRepo := repos.NewHabitsRepoDynamo(db)
	streaksRepo := repos.NewStreaksRepoDynamo(db)
	motivationService := services.NewMotivationService(services.NewMotivationServiceInput{
		Kudos:        nil,
		Affirmations: nil,
	})
	userService := services.NewUserService(usersRepo)
	habitService := services.NewHabitService(habitsRepo, streaksRepo, &common.RealClock{})
	handler := internal.NewHandler(bot, userService, habitService, motivationService)

	users, err := userService.FindAll()
	if err != nil {
		fmt.Println("can't get users")
	}

	for i, u := range users {
		if i%100 == 0 {
			time.Sleep(1000)
		}
		if u.ChatId != 0 {
			activeHabits, err := habitService.GetAllActive(u.Id)
			if err != nil {
				errors++
				continue
			}
			if len(activeHabits) == 0 {
				skipped++
				continue
			}
			err = sendKeyboard(u.ChatId, u, handler)
			if err != nil {
				errors++
			}
		}
	}

	fmt.Println("users count", len(users))
	fmt.Println("sent ", sent)
	fmt.Println("errors ", errors)
}

func sendKeyboard(chatId int64, user *repos.UserEntity, handler *internal.Handler) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
			errors++
		} else {
			sent++
		}
	}()
	return handler.CommandSendKeyboard(chatId, user)
}
