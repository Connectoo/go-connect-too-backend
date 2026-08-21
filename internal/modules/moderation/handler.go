package moderation

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/reports"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/reviews"
	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/pagination"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
)

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.Reviews().ListForAdmin(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeReviewError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Reviews loaded", res)
}

func (h *Handler) approveReview(w http.ResponseWriter, r *http.Request) {
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid review id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Reviews().Approve(r.Context(), reviewID)
	if err != nil {
		h.writeReviewError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Review approved", res)
}

func (h *Handler) hideReview(w http.ResponseWriter, r *http.Request) {
	reviewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid review id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Reviews().Hide(r.Context(), reviewID)
	if err != nil {
		h.writeReviewError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Review hidden", res)
}

func (h *Handler) listReports(w http.ResponseWriter, r *http.Request) {
	page := pagination.Parse(r.URL.Query())
	res, err := h.svc.Reports().ListForAdmin(r.Context(), r.URL.Query().Get("status"), page)
	if err != nil {
		h.writeReportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Reports loaded", res)
}

func (h *Handler) resolveReport(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid report id", sharederrors.CodeValidationError)
		return
	}

	res, err := h.svc.Reports().Resolve(r.Context(), reportID)
	if err != nil {
		h.writeReportError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, "Report resolved", res)
}

func (h *Handler) writeReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, reviews.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Review not found", sharederrors.CodeNotFound)
	case errors.Is(err, reviews.ErrInvalidStatus):
		response.Error(w, http.StatusConflict, "Review cannot be approved", sharederrors.CodeConflict)
	default:
		if h.log != nil {
			h.log.Error("moderation review request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}

func (h *Handler) writeReportError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, reports.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Report not found", sharederrors.CodeNotFound)
	default:
		if h.log != nil {
			h.log.Error("moderation report request failed", slog.String("error", err.Error()))
		}
		response.Error(w, http.StatusInternalServerError, "Something went wrong", sharederrors.CodeInternalError)
	}
}
