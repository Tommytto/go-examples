package repos

import (
	"errors"
	"github.com/Tommytto/habit-bot/internal/config"
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"time"
)

type HabitThoughtEntity struct {
	Text      string    `dynamo:"text" json:"text"`
	CreatedAt time.Time `dynamo:"created_at,set" json:"created_at"`
}
type HabitEntity struct {
	Id       string                `dynamo:"id" json:"id"`
	UserId   UserId                `dynamo:"user_id" json:"user_id"`
	Name     string                `dynamo:"name" json:"name"`
	Archived bool                  `dynamo:"archived" json:"archived"`
	Thoughts []*HabitThoughtEntity `dynamo:"thoughts,set" json:"thoughts"`
}

//go:generate mockgen -destination=../mocks/mock_habit_repo.go -package=mocks . HabitsRepo
type HabitsRepo interface {
	Create(entity *HabitEntity) (*HabitEntity, error)
	GetAll(userId UserId) ([]*HabitEntity, error)
	Get(habitId string) (*HabitEntity, error)
	AddThought(habitId string, text string) error
	UpdateOne(habitId string, toUpdate map[string]interface{}) error
}

type HabitsRepoDynamo struct {
	Table dynamo.Table
}

func NewHabitsRepoDynamo(db *dynamo.DB) HabitsRepo {
	return &HabitsRepoDynamo{Table: db.Table(config.HabitsTableName)}
}

func (h *HabitsRepoDynamo) Create(entity *HabitEntity) (*HabitEntity, error) {
	entity.Id = uuid.New().String()
	if err := h.Table.Put(entity).Run(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (h *HabitsRepoDynamo) GetAll(userId UserId) ([]*HabitEntity, error) {
	var result []*HabitEntity
	err := h.Table.Scan().Filter("'user_id' = ?", userId).All(&result)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (h *HabitsRepoDynamo) Get(habitId string) (*HabitEntity, error) {
	var result *HabitEntity
	err := h.Table.Get("id", habitId).One(&result)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, ers.ErrNotFound
		}
		return nil, err
	}

	return result, nil
}

func (h *HabitsRepoDynamo) AddThought(habitId string, text string) error {
	if err := h.Table.
		Update("id", habitId).
		SetIfNotExists("thoughts", []*HabitThoughtEntity{}).
		Run(); err != nil {
		return err
	}

	if err := h.Table.
		Update("id", habitId).
		Append("thoughts", []*HabitThoughtEntity{{
			Text:      text,
			CreatedAt: time.Now(),
		}}).
		Run(); err != nil {
		return err
	}

	return nil
}

func (h *HabitsRepoDynamo) UpdateOne(habitId string, toUpdate map[string]interface{}) error {
	update := h.Table.Update("id", habitId)
	for path, value := range toUpdate {
		if path == "id" {
			continue
		}
		update = update.Set(path, value)
	}

	return update.Run()
}
