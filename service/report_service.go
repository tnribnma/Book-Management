package service

import (
	"context"
	"fmt"
	"book-management/models"
	"book-management/repository"
)

type ReportService interface {
	GetDashboard(ctx context.Context) (*models.DashboardStats, error)
	GetPopularBooks(ctx context.Context, limit int) ([]models.Book, error)
	GetSystemSummary(ctx context.Context) (*models.SystemSummary, error)
}

type reportService struct {
	reportRepo    repository.ReportRepository
	bookRepo      repository.BookRepository
	borrowingRepo repository.BorrowingRepository
}

func NewReportService(
	reportRepo repository.ReportRepository,
	bookRepo repository.BookRepository,
	borrowingRepo repository.BorrowingRepository,
) ReportService {
	return &reportService{
		reportRepo:    reportRepo,
		bookRepo:      bookRepo,
		borrowingRepo: borrowingRepo,
	}
}

func (s *reportService) GetDashboard(ctx context.Context) (*models.DashboardStats, error) {
	stats, err := s.reportRepo.GetDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard stats: %w", err)
	}
	return stats, nil
}

func (s *reportService) GetPopularBooks(ctx context.Context, limit int) ([]models.Book, error) {
	if limit <= 0 || limit > 20 {
		limit = 10 
	}

	books, err := s.reportRepo.GetPopularBooks(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular books: %w", err)
	}

	return books, nil
}

func (s *reportService) GetSystemSummary(ctx context.Context) (*models.SystemSummary, error) {
	summary := &models.SystemSummary{}

	totalBooks, err := s.reportRepo.GetTotalBooksCount(ctx)
	if err != nil {
		return nil, err
	}
	summary.TotalBooks = totalBooks

	activeBorrowings, err := s.reportRepo.GetActiveBorrowingsCount(ctx)
	if err != nil {
		return nil, err
	}
	summary.ActiveBorrowings = activeBorrowings

	overdue, err := s.reportRepo.GetOverdueCount(ctx)
	if err != nil {
		return nil, err
	}
	summary.OverdueBorrowings = overdue

	totalUsers, err := s.reportRepo.GetTotalUsersCount(ctx)
	if err != nil {
		return nil, err
	}
	summary.TotalUsers = totalUsers

	dashboard, err := s.reportRepo.GetDashboardStats(ctx)
	if err == nil {
		summary.AvailableBooks = dashboard.AvailableBooks
	}

	return summary, nil
}