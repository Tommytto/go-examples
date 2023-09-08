package services

import (
	"errors"
	"fmt"
	"github.com/Tommytto/habit-bot/internal/common"
	"github.com/Tommytto/habit-bot/internal/ers"
	"github.com/Tommytto/habit-bot/internal/repos"
	"log"
	"sort"
	"time"
)

type HabitService struct {
	habitRepo   repos.HabitsRepo
	streaksRepo repos.StreaksRepo
	clock       common.Clock
}

func NewHabitService(habitRepo repos.HabitsRepo, streaksRepo repos.StreaksRepo, clock common.Clock) *HabitService {
	return &HabitService{habitRepo: habitRepo, streaksRepo: streaksRepo, clock: clock}
}

type CreateHabitInput struct {
	UserId        repos.UserId
	Name          string
	DaysCompleted int
	CreatedAt     time.Time
}

func (h *HabitService) CreateHabit(input *CreateHabitInput) (*repos.HabitEntity, error) {
	if input.Name == "" || input.UserId == "" {
		return nil, ers.ErrBadInput
	}
	habit, err := h.habitRepo.Create(&repos.HabitEntity{
		UserId: input.UserId,
		Name:   input.Name,
	})
	if err != nil {
		log.Print("create habit problem", err)
		return nil, err
	}

	if input.DaysCompleted < 1 {
		return habit, nil
	}

	t := h.clock.Now()
	createDate := input.CreatedAt

	log.Printf("\nserver time: %v\n, user time: %v", t, createDate)
	streakEntity := &repos.StreakEntity{
		HabitId:   habit.Id,
		StartDate: createDate.AddDate(0, 0, -input.DaysCompleted),
		EndDate:   createDate.AddDate(0, 0, -1),
	}

	_, err = h.streaksRepo.Create(streakEntity)
	if err != nil {
		log.Print("create streak problem", err)
		return nil, err
	}
	return habit, nil
}

// GetAll returns even archived entities
func (h *HabitService) GetAll(userId repos.UserId) ([]*repos.HabitEntity, error) {
	habits, err := h.habitRepo.GetAll(userId)
	if err != nil {
		fmt.Println("can't get all habits for user", userId)
		return nil, err
	}

	return habits, nil
}

func (h *HabitService) GetAllActive(userId repos.UserId) ([]*repos.HabitEntity, error) {
	habits, err := h.GetAll(userId)
	if err != nil {
		fmt.Println("can't get all habits for user", userId)
		return nil, err
	}

	var activeHabits []*repos.HabitEntity

	for _, habit := range habits {
		if !habit.Archived {
			activeHabits = append(activeHabits, habit)
		}
	}

	return activeHabits, nil
}

func (h *HabitService) GetOne(habitId string) (*repos.HabitEntity, error) {
	habit, err := h.habitRepo.Get(habitId)
	if err != nil {
		log.Printf("habit not found %v", err)
		return nil, err
	}

	return habit, nil
}

var ErrAlreadyCompleted = errors.New("habit is already completed today")

// CompleteToday todo validate habit owner
func (h *HabitService) CompleteToday(habitId string, completedAt time.Time) (*repos.StreakEntity, error) {
	streak, err := h.streaksRepo.GetCurrentStreak(habitId, completedAt)
	if errors.Is(err, ers.ErrNotFound) {
		streak, err = h.streaksRepo.Create(&repos.StreakEntity{
			HabitId:   habitId,
			StartDate: completedAt,
			EndDate:   completedAt,
		})
		if err != nil {
			fmt.Println("can't create new streak", err)
			return nil, err
		}
		return streak, nil
	} else if err != nil {
		log.Print("can't find current streak: ", err)
		return nil, err
	}

	y1, m1, d1 := completedAt.Date()
	y2, m2, d2 := streak.EndDate.Date()
	if y1 == y2 && m1 == m2 && d1 == d2 {
		return nil, ErrAlreadyCompleted
	}

	newEndDate := completedAt
	err = h.streaksRepo.UpdateOne(streak.HabitId, streak.StartDate, map[string]interface{}{
		"end_date": newEndDate,
	})
	if err != nil {
		log.Print("can't update streak for today: ", err)
		return nil, err
	}
	streak.EndDate = newEndDate

	return streak, nil
}

func (h *HabitService) GetCurrentStreakLength(habitId string, when time.Time) (int, error) {
	streak, err := h.streaksRepo.GetCurrentStreak(habitId, when)
	if err != nil {
		if errors.Is(err, ers.ErrNotFound) {
			return 0, nil
		}
		log.Print("can't find current streak: ", err)
		return 0, err
	}

	return GetCalendarDaysDiff(streak.StartDate, streak.EndDate), nil
}

func GetCalendarDaysDiff(startDate time.Time, endDate time.Time) int {
	hoursDiff := endDate.Sub(startDate).Hours()
	// started and ended equal times mean user completed habit once
	if hoursDiff == 0 {
		return 1
	}

	// add one because
	// startDate=12.12.2012 12:00
	// endDate=13.12.2012 13:23
	// means that he completed habit yesterday at 12:00 and todat at 13:23,
	// but hours diff is only 25:23 hours (1 full day)
	fullDaysDiff := int(hoursDiff / 24)

	// user completed habit yesterday at 12, but today at 9
	// diff is less then 24 hours, so fullDays equal 0, 1 day added to fix that
	daysDiff := fullDaysDiff + 1

	// user completed yesterday at 12, so streak equals 1
	// user completed today at 9, so streak equals 2
	// We have to compare if days are different, we need to add 1 more day to streak
	// if less 24 hours past
	if endDate.Day() != startDate.Day() {
		if endDate.Hour() < startDate.Hour() ||
			endDate.Minute() < startDate.Minute() ||
			endDate.Second() < startDate.Second() ||
			endDate.Nanosecond() < startDate.Nanosecond() {
			daysDiff += 1
		}
	}

	return daysDiff
}

func getLastThought(habit *repos.HabitEntity) *repos.HabitThoughtEntity {
	if habit.Thoughts == nil {
		return nil
	}

	sort.Slice(habit.Thoughts, func(i, j int) bool {
		return habit.Thoughts[i].CreatedAt.After(habit.Thoughts[j].CreatedAt)
	})

	return habit.Thoughts[0]
}

func (h *HabitService) NeedAskThought(habitId string, when time.Time) bool {
	var lastThought *repos.HabitThoughtEntity
	if habit, err := h.GetOne(habitId); err != nil {
		log.Print("need ask though problem: ", err)
		return false
	} else {
		lastThought = getLastThought(habit)
	}

	// if today wrote already
	if lastThought != nil && common.SameDay(when, lastThought.CreatedAt) {
		return false
	}

	streakLength, err := h.GetCurrentStreakLength(habitId, when)
	if err != nil {
		log.Print("need ask though problem: ", err)
		return false
	}

	if streakLength == 0 {
		return false
	}

	if streakLength == 1 ||
		streakLength == 3 ||
		streakLength%7 == 0 {
		return true
	}

	return false
}

func (h *HabitService) AddThought(habitId string, text string) error {
	if err := h.habitRepo.AddThought(habitId, text); err != nil {
		return err
	}

	return nil
}

func (h *HabitService) ToggleArchive(habitId string) error {
	habit, err := h.GetOne(habitId)
	if err != nil {
		return err
	}
	err = h.habitRepo.UpdateOne(habitId, map[string]interface{}{
		"archived": !habit.Archived,
	})
	if err != nil {
		return err
	}

	return nil
}
