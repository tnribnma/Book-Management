package service

import (
	"math"
	"time"
	"book-management/models"
)

type FineStrategy interface {
	Calculate(b models.Borrowing) float64
}

type DailyFlatFine struct {
	RatePerDay float64
}

func NewDailyFlatFine(rate float64) *DailyFlatFine {
	return &DailyFlatFine{RatePerDay: rate}
}

func (f *DailyFlatFine) Calculate(b models.Borrowing) float64 {
	if b.Status != "borrowed" || !b.DueDate.Before(time.Now()) {
		return 0
	}
	daysOverdue := math.Ceil(time.Since(b.DueDate).Hours() / 24)
	return daysOverdue * f.RatePerDay
}

type TieredFine struct {
	GraceDays   float64
	BaseRate    float64
	PenaltyRate float64
}

func NewTieredFine(graceDays int, baseRate, penaltyRate float64) *TieredFine {
	return &TieredFine{
		GraceDays:   float64(graceDays),
		BaseRate:    baseRate,
		PenaltyRate: penaltyRate,
	}
}

func (f *TieredFine) Calculate(b models.Borrowing) float64 {
	if b.Status != "borrowed" || !b.DueDate.Before(time.Now()) {
		return 0
	}
	daysOverdue := math.Ceil(time.Since(b.DueDate).Hours() / 24)
	if daysOverdue <= f.GraceDays {
		return daysOverdue * f.BaseRate
	}
	graceAmount := f.GraceDays * f.BaseRate
	penaltyAmount := (daysOverdue - f.GraceDays) * f.PenaltyRate
	return graceAmount + penaltyAmount
}