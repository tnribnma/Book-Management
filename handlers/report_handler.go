package handlers

import (
	"database/sql"
	"net/http"

	"book-management/service"
)

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(db *sql.DB) *ReportHandler {
	return &ReportHandler{service: service.NewReportService(db)}
}

func (h *ReportHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate report")
		return
	}
	JSON(w, http.StatusOK, report)
}

func (h *ReportHandler) GetOverdueBooks(w http.ResponseWriter, r *http.Request) {
	overdue, err := h.service.GetOverdueBooks(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, overdue)
}