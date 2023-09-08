package repos

import (
	"errors"
	"fmt"
	"github.com/Tommytto/habit-bot/internal/config"
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"log"
	"time"
)

const usersTablePK = "id"
const usersTableTelegramIndex = "telegram_id-index"
const usersTableTelegramIndexPK = "telegram_id"

type UserId = string

type UserEntity struct {
	Id               UserId `dynamo:"id" json:"id"`
	TelegramId       int64  `dynamo:"telegram_id" index:"telegram_id-index,hash" json:"telegramId"`
	Step             string `dynamo:"step" json:"step"`
	ChatId           int64  `dynamo:"chat_id" json:"chat_id"`
	FirstName        string `dynamo:"first_name" json:"first_name"`
	LastName         string `dynamo:"last_name" json:"last_name"`
	TelegramUsername string `dynamo:"telegram_username" json:"telegram_username"`
}

//go:generate mockgen -destination=../mocks/mock_users_repo.go -package=mocks . UsersRepo
type UsersRepo interface {
	FindOneByTelegramId(int64) (*UserEntity, error)
	FindAll() ([]*UserEntity, error)
	UpdateOne(UserId, map[string]interface{}) error
	Create(UserEntity) (*UserEntity, error)
	GetAndSetAffirmationToday(entity *UserEntity, when time.Time) bool
	AffirmationWasSentToday(entity *UserEntity, when time.Time) bool
}

type UsersRepoDynamo struct {
	db                 *dynamo.DB
	usersTable         dynamo.Table
	telegramUsersTable dynamo.Table
}

func NewUsersRepoDynamo(db *dynamo.DB) UsersRepo {
	return &UsersRepoDynamo{
		db:                 db,
		usersTable:         db.Table(config.UsersTableName),
		telegramUsersTable: db.Table(config.TelegramUsersTableName),
	}
}

func (u *UsersRepoDynamo) Create(user UserEntity) (*UserEntity, error) {
	user.Id = uuid.New().String()

	tx := u.db.WriteTx()

	// for telegram_id uniqueness
	tx.Put(u.usersTable.Put(map[string]interface{}{
		usersTablePK: fmt.Sprintf("telegram_id#%v", user.TelegramId),
	}))
	tx.Put(u.usersTable.Put(user))

	if err := tx.Run(); err != nil {
		return nil, fmt.Errorf("can't create user %s", err)
	}

	return &user, nil
}

func (u *UsersRepoDynamo) FindOneByTelegramId(telegramId int64) (*UserEntity, error) {
	var err error
	var user *UserEntity
	err = u.usersTable.Get(usersTableTelegramIndexPK, telegramId).Index(usersTableTelegramIndex).One(&user)
	if errors.Is(err, dynamo.ErrNotFound) {
		log.Print("user not found", err)
		return nil, ers.ErrNotFound
	} else if err != nil {
		log.Print("find user problem: ", err)
		return nil, err
	}

	return user, nil
}

func (u *UsersRepoDynamo) UpdateOne(userId UserId, toUpdate map[string]interface{}) error {
	update := u.usersTable.Update(usersTablePK, userId)
	for path, value := range toUpdate {
		update = update.Set(path, value)
	}

	return update.Run()
}

func (u *UsersRepoDynamo) FindAll() ([]*UserEntity, error) {
	var users []*UserEntity
	err := u.usersTable.Scan().Filter("attribute_exists(chat_id)", nil).All(&users)
	if err != nil {
		log.Print("can't find all users", err)
		return nil, err
	}

	return users, nil
}

var affirmationsMap = make(map[string]bool)

// GetAndSetAffirmationToday refactor and make persistent storage, move store logic to repo
func (u *UsersRepoDynamo) GetAndSetAffirmationToday(user *UserEntity, when time.Time) bool {
	key, date := getAffirmationMapKey(user.Id, when)
	// renew map
	_, ok := affirmationsMap[date]
	if !ok {
		affirmationsMap = make(map[string]bool)
		affirmationsMap[date] = true
	}

	was := affirmationsMap[key]
	affirmationsMap[key] = true
	return was
}

func (u *UsersRepoDynamo) AffirmationWasSentToday(user *UserEntity, when time.Time) bool {
	key, _ := getAffirmationMapKey(user.Id, when)
	return affirmationsMap[key]
}

func getAffirmationMapKey(userId UserId, when time.Time) (string, string) {
	date := when.Format("02.04.06")
	return date + "_" + userId, date
}
