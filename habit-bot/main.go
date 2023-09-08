package main

import (
	"github.com/Tommytto/habit-bot/internal"
	"github.com/Tommytto/habit-bot/internal/aws"
	"github.com/Tommytto/habit-bot/internal/common"
	"github.com/Tommytto/habit-bot/internal/repos"
	"github.com/Tommytto/habit-bot/internal/services"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	bot.Debug = debug

	internal.SetupCommands(bot)

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	kudos, err := services.LoadKudos()
	if err != nil {
		panic("can't load kudos file")
	}
	affirmations, err := services.LoadAffirmations()
	if err != nil {
		panic("can't load affirmations file")
	}

	awsWrapper := aws.NewAwsWrapper()
	db := awsWrapper.GetDynamo()
	userRepo := repos.NewUsersRepoDynamo(db)
	habitRepo := repos.NewHabitsRepoDynamo(db)
	streaksRepo := repos.NewStreaksRepoDynamo(db)
	motivationService := services.NewMotivationService(services.NewMotivationServiceInput{
		Kudos:        kudos,
		Affirmations: affirmations,
	})
	userService := services.NewUserService(userRepo)
	habitService := services.NewHabitService(habitRepo, streaksRepo, common.RealClock{})
	handler := internal.NewHandler(bot, userService, habitService, motivationService)

	//httpHandler := internal.NewHttpHandler()
	//http.Handle("/", http.HandlerFunc(httpHandler.Healthcheck))

	//go func() {
	//	err = http.ListenAndServe(":8080", nil)
	//	if err != nil {
	//		log.Fatalf("fatal error in ListenAndServe: %v", err)
	//	}
	//}()
	for update := range updates {
		handler.HandleUpdate(&update)
	}
}
