package repos

import (
	"errors"
	"fmt"
	"github.com/Tommytto/habit-bot/internal/config"
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/guregu/dynamo"
	"log"
	"sort"
	"time"
)

type StreakEntity struct {
	HabitId   string    `dynamo:"habit_id" json:"habit_id"`
	StartDate time.Time `dynamo:"start_date" json:"start_date"`
	EndDate   time.Time `dynamo:"end_date" json:"end_date"`
}

//go:generate mockgen -destination=../mocks/mock_streaks_repo.go -package=mocks . StreaksRepo
type StreaksRepo interface {
	Create(entity *StreakEntity) (*StreakEntity, error)
	Get(streakId string) (*StreakEntity, error)
	GetAll(habitId string) ([]*StreakEntity, error)
	GetCurrentStreak(habitId string, when time.Time) (*StreakEntity, error)
	UpdateOne(habitId string, startDate time.Time, toUpdate map[string]interface{}) error
}

type StreaksRepoDynamo struct {
	Table dynamo.Table
}

func NewStreaksRepoDynamo(db *dynamo.DB) StreaksRepo {
	return &StreaksRepoDynamo{Table: db.Table(config.StreaksTableName)}
}

func (s *StreaksRepoDynamo) Create(entity *StreakEntity) (*StreakEntity, error) {
	if err := s.Table.Put(entity).Run(); err != nil {
		return nil, fmt.Errorf("can't create habit %s", err)
	}

	return entity, nil
}

func (s *StreaksRepoDynamo) Get(streakId string) (*StreakEntity, error) {
	var streak *StreakEntity
	err := s.Table.Get("streak_id", streakId).One(&streak)
	if err != nil {
		if err == dynamo.ErrNotFound {
			return nil, ers.ErrNotFound
		}
		log.Println("can't search telegram user by telegram_id", err)
		return nil, err
	}

	return streak, nil
}

func (s *StreaksRepoDynamo) GetAll(habitId string) ([]*StreakEntity, error) {
	var result []*StreakEntity
	err := s.Table.Scan().Filter("'habit_id' = ?", habitId).All(&result)
	if err != nil {
		fmt.Println("can't get all habits")
		return nil, err
	}

	return result, nil
}

// GetCurrentStreak return streak.end_date < "yesterday start of the day"
// because person probably have not clicked yet for today
func (s *StreaksRepoDynamo) GetCurrentStreak(habitId string, when time.Time) (*StreakEntity, error) {
	sDate := when.AddDate(0, 0, -1)
	sDateStart := time.Date(sDate.Year(), sDate.Month(), sDate.Day(), 0, 0, 0, 0, sDate.Location())

	var lastStreaks []*StreakEntity
	err := s.Table.
		Scan().
		Filter("'habit_id' = ?", habitId).
		Filter("'end_date' > ?", sDateStart).
		All(&lastStreaks)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return nil, ers.ErrNotFound
		}
		fmt.Println("can't get habit")
		return nil, err
	}

	if len(lastStreaks) == 0 {
		return nil, ers.ErrNotFound
	}

	sort.Slice(lastStreaks, func(i, j int) bool {
		return lastStreaks[i].EndDate.Nanosecond() < lastStreaks[j].EndDate.Nanosecond()
	})
	lastStreak := lastStreaks[0]
	return lastStreak, nil
}

func (s *StreaksRepoDynamo) UpdateOne(habitId string, startDate time.Time, toUpdate map[string]interface{}) error {
	update := s.Table.Update("habit_id", habitId).Range("start_date", startDate)
	for path, value := range toUpdate {
		if path == "streak_id" {
			continue
		}
		update = update.Set(path, value)
	}

	return update.Run()
}
