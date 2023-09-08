package services

import (
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/Tommytto/habit-bot/internal/mocks"
	"github.com/Tommytto/habit-bot/internal/repos"
	"github.com/golang/mock/gomock"
	"testing"
)

func newUserService(t *testing.T) (*UserService, *mocks.MockUsersRepo) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUsersRepo(ctrl)
	return NewUserService(repo), repo
}

func TestUserService_FindOrCreateByTelegramId_Found(t *testing.T) {
	us, repo := newUserService(t)

	createUserInput := CreateUserInput{
		TelegramId: 111,
		ChatId:     111,
	}
	repo.
		EXPECT().
		FindOneByTelegramId(createUserInput.TelegramId).
		Return(&repos.UserEntity{Id: "id", ChatId: 111, TelegramId: 111}, nil)

	user, err := us.FindOrCreateByTelegramId(&createUserInput)

	if err != nil ||
		user.TelegramId != 111 ||
		user.ChatId != 111 ||
		user.Id != "id" {
		t.Fatalf("user not equal")
	}
}

func TestUserService_FindOrCreateByTelegramId_Create(t *testing.T) {
	us, repo := newUserService(t)
	createUserInput := CreateUserInput{
		TelegramId: 111,
		ChatId:     111,
	}
	repo.
		EXPECT().
		FindOneByTelegramId(int64(111)).
		Return(nil, ers.ErrNotFound)
	repo.
		EXPECT().
		Create(
			repos.UserEntity{
				TelegramId: createUserInput.TelegramId,
				ChatId:     createUserInput.ChatId,
			},
		).
		Return(&repos.UserEntity{Id: "id", TelegramId: createUserInput.TelegramId, ChatId: 111}, nil)

	user, err := us.FindOrCreateByTelegramId(&createUserInput)
	if err != nil {
		t.Fatalf("err %s", err)
	}
	if user.TelegramId != 111 ||
		user.ChatId != 111 ||
		user.Id != "id" {
		t.Fatalf("user not equal")
	}
}
