package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtQuestion struct {
	svc *service.CbtQuestion
}

func NewCbtQuestion(svc *service.CbtQuestion) *CbtQuestion { return &CbtQuestion{svc: svc} }

func (h *CbtQuestion) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtQuestion) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SubjectID    string `json:"subject_id"`
		Code         string `json:"code"`
		QuestionText string `json:"question_text"`
		OptionA      string `json:"option_a"`
		OptionB      string `json:"option_b"`
		OptionC      string `json:"option_c"`
		OptionD      string `json:"option_d"`
		OptionE      string `json:"option_e"`
		AnswerKey    string `json:"answer_key"`
		Explanation  string `json:"explanation"`
		Difficulty   string `json:"difficulty"`
		Status       string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		api.BadRequest(w, "subject_id invalid")
		return
	}
	row, err := h.svc.Create(r.Context(), db.CreateCbtQuestionParams{
		SubjectID:    subjectID,
		Code:         body.Code,
		QuestionText: body.QuestionText,
		OptionA:      body.OptionA,
		OptionB:      body.OptionB,
		OptionC:      body.OptionC,
		OptionD:      body.OptionD,
		OptionE:      body.OptionE,
		AnswerKey:    body.AnswerKey,
		Explanation:  body.Explanation,
		Difficulty:   db.CbtQuestionDifficultyEnum(body.Difficulty),
		Status:       db.CbtQuestionStatusEnum(body.Status),
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *CbtQuestion) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}
