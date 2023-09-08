package services

import (
	"errors"
	"fmt"
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/Tommytto/habit-bot/internal/repos"
	"github.com/guregu/dynamo"
	"log"
	"time"
)

type UserService struct {
	userRepo repos.UsersRepo
}

func NewUserService(userRepo repos.UsersRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserInput struct {
	TelegramId       int64
	ChatId           int64
	FirstName        string
	LastName         string
	TelegramUsername string
}

func (us *UserService) CreateByTelegramId(input *CreateUserInput) (*repos.UserEntity, error) {
	user, err := us.userRepo.Create(repos.UserEntity{
		TelegramId:       input.TelegramId,
		ChatId:           input.ChatId,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		TelegramUsername: input.TelegramUsername,
	})
	if err != nil {
		log.Print("can't create user", err)
		return nil, err
	}
	return user, err
}

func (us *UserService) FindOneByTelegramId(telegramId int64) (*repos.UserEntity, error) {
	existedUser, err := us.userRepo.FindOneByTelegramId(telegramId)
	if err != nil {
		return nil, err
	}

	return existedUser, nil
}

func (us *UserService) FindAll() ([]*repos.UserEntity, error) {
	users, err := us.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (us *UserService) FindOrCreateByTelegramId(input *CreateUserInput) (*repos.UserEntity, error) {
	user, err := us.FindOneByTelegramId(input.TelegramId)
	if errors.Is(err, ers.ErrNotFound) {
		user, err = us.CreateByTelegramId(input)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		log.Print("user not found")
		return nil, err
	}

	// todo remove this when all users fill information or not??
	// probably it can be useful to update information
	updateMap := make(map[string]interface{})
	if input.FirstName != "" && user.FirstName != input.FirstName {
		updateMap["first_name"] = input.FirstName
	}
	if input.LastName != "" && user.LastName != input.LastName {
		updateMap["last_name"] = input.LastName
	}
	if input.TelegramUsername != "" && user.TelegramUsername != input.TelegramUsername {
		updateMap["telegram_username"] = input.TelegramUsername
	}
	if input.ChatId != 0 && user.ChatId != input.ChatId {
		updateMap["chat_id"] = input.ChatId
	}

	if len(updateMap) > 0 {
		fmt.Println("input", input)
		fmt.Println("user", user)
		fmt.Println("update map", updateMap)
		err := us.userRepo.UpdateOne(user.Id, updateMap)
		if err != nil && err != dynamo.ErrNotFound {
			log.Print("can't update chat_id", err)
			return nil, err
		}
	}

	return user, nil
}

func (us *UserService) SetStep(user *repos.UserEntity, step string) error {
	err := us.userRepo.UpdateOne(user.Id, map[string]interface{}{
		"step": step,
	})
	if err != nil {
		log.Print("set step problem", err)
		return err
	}

	return nil
}

func (us *UserService) GetStep(user *repos.UserEntity) (string, error) {
	if user != nil {
		return user.Step, nil
	}

	return "", fmt.Errorf("No user")
}

var affirmationsMap = make(map[string]bool)

func (us *UserService) GetAndSetAffirmationToday(user *repos.UserEntity, when time.Time) bool {
	return us.userRepo.GetAndSetAffirmationToday(user, when)
}

func (us *UserService) AffirmationWasSentToday(user *repos.UserEntity, when time.Time) bool {
	return us.userRepo.AffirmationWasSentToday(user, when)
}

func (us *UserService) ResetStep(user *repos.UserEntity) error {
	return us.SetStep(user, "")
}
