package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var stripHTMLTags = regexp.MustCompile(`(?s)<[^>]*>`)
var stripDangerousBlockPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
	regexp.MustCompile(`(?is)<iframe[^>]*>.*?</iframe>`),
	regexp.MustCompile(`(?is)<object[^>]*>.*?</object>`),
	regexp.MustCompile(`(?is)<embed[^>]*>.*?</embed>`),
}
var stripEventHandlers = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*(".*?"|'.*?'|[^\s>]+)`)
var stripDangerousURLs = regexp.MustCompile(`(?i)\s(href|src)\s*=\s*(['"])\s*javascript:[^'"]*['"]`)
var importAnswerTokenSeparators = regexp.MustCompile(`[,\|;/+\s]+`)
var importMatchingPairSeparators = regexp.MustCompile(`[;,|]+`)

type cbtQuestionStore interface {
	ListCbtQuestions(ctx context.Context, arg db.ListCbtQuestionsParams) ([]db.ListCbtQuestionsRow, error)
	ListCbtQuestionsFiltered(ctx context.Context, arg db.ListCbtQuestionsFilteredParams) ([]db.ListCbtQuestionsFilteredRow, error)
	CountCbtQuestionsFiltered(ctx context.Context, arg db.CountCbtQuestionsFilteredParams) (int64, error)
	GetCbtQuestionSummaryCounts(ctx context.Context, arg db.GetCbtQuestionSummaryCountsParams) (db.GetCbtQuestionSummaryCountsRow, error)
	ListCbtQuestionSummaryBySubject(ctx context.Context, arg db.ListCbtQuestionSummaryBySubjectParams) ([]db.ListCbtQuestionSummaryBySubjectRow, error)
	ListCbtQuestionSummaryByCognitiveLevel(ctx context.Context, arg db.ListCbtQuestionSummaryByCognitiveLevelParams) ([]db.ListCbtQuestionSummaryByCognitiveLevelRow, error)
	ListCbtQuestionSummaryRecent(ctx context.Context, arg db.ListCbtQuestionSummaryRecentParams) ([]db.ListCbtQuestionSummaryRecentRow, error)
	ListCbtEventMembersByUser(ctx context.Context, userID pgtype.UUID) ([]db.CbtEventMember, error)
	ListCbtEventMembersByUsername(ctx context.Context, username string) ([]db.CbtEventMember, error)
	ListCbtQuestionStemTextsBySubject(ctx context.Context, subjectID pgtype.UUID) ([]db.ListCbtQuestionStemTextsBySubjectRow, error)
	GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
	GetCbtQuestionDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error)
	CreateCbtQuestion(ctx context.Context, arg db.CreateCbtQuestionParams) (db.CbtQuestion, error)
	UpdateCbtQuestion(ctx context.Context, arg db.UpdateCbtQuestionParams) (db.CbtQuestion, error)
	DeleteCbtQuestion(ctx context.Context, id pgtype.UUID) error
	CreateCbtQuestionAuditLog(ctx context.Context, arg db.CreateCbtQuestionAuditLogParams) (db.CbtQuestionAuditLog, error)
	ListCbtQuestionTimeline(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAuditLog, error)
}

type CbtQuestion struct {
	q cbtQuestionStore
}

func NewCbtQuestion(q *db.Queries) *CbtQuestion { return &CbtQuestion{q: q} }

type QuestionOption struct {
	Label        string `json:"label"`
	Text         string `json:"text,omitempty"`
	HTML         string `json:"html,omitempty"`
	Latex        string `json:"latex,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	MatchLabel   string `json:"match_label,omitempty"`
	MatchText    string `json:"match_text,omitempty"`
	MatchHTML    string `json:"match_html,omitempty"`
	IsDistractor bool   `json:"is_distractor,omitempty"`
}

type SaveCbtQuestionInput struct {
	ID               pgtype.UUID
	EventID          pgtype.UUID
	SubjectID        pgtype.UUID
	AuthoringMode    string
	Code             string
	QuestionText     string
	QuestionType     string
	Options          []QuestionOption
	OptionA          string
	OptionB          string
	OptionC          string
	OptionD          string
	OptionE          string
	AnswerKey        string
	Explanation      string
	Difficulty       db.CbtQuestionDifficultyEnum
	Status           db.CbtQuestionStatusEnum
	StemHTML         string
	StemLatex        string
	StimulusHTML     string
	StimulusLatex    string
	ExplanationHTML  string
	RubricHTML       string
	AcademicPhase    string
	GradeLevel       pgtype.Int2
	CPRef            string
	TPRef            string
	KDRef            string
	IndicatorRef     string
	MaterialTopic    string
	CognitiveLevel   string
	HotsFlag         bool
	MediaAssetIDs    []string
	WorkflowStatus   string
	AuthorUsername   string
	ReviewerUsername string
	ApproverUsername string
	WriterNotes      string
	ReviewNotes      string
	Actor            CbtQuestionActor
}

type CbtQuestionActor struct {
	UserID   pgtype.UUID
	Username string
	Roles    []string
}

func (a CbtQuestionActor) HasRole(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, role := range a.Roles {
		if strings.TrimSpace(role) == target {
			return true
		}
	}
	return false
}

func (a CbtQuestionActor) IsAdmin() bool {
	return a.HasRole("admin")
}

func normalizeCbtQuestionActor(actor CbtQuestionActor) CbtQuestionActor {
	actor.Username = strings.TrimSpace(actor.Username)
	return actor
}

func inputActor(input SaveCbtQuestionInput) CbtQuestionActor {
	actor := normalizeCbtQuestionActor(input.Actor)
	if actor.Username == "" {
		actor.Username = strings.TrimSpace(input.AuthorUsername)
	}
	return actor
}

func (s *CbtQuestion) List(ctx context.Context) ([]db.ListCbtQuestionsRow, error) {
	return s.q.ListCbtQuestions(ctx, db.ListCbtQuestionsParams{IsAdmin: true})
}

type ListCbtQuestionsInput struct {
	EventID        pgtype.UUID
	QuestionScope  string
	SubjectID      pgtype.UUID
	AuthorUsername string
	WorkflowStatus string
	Status         string
	QuestionType   string
	HotsFilter     string
	RevisionSource string
	SearchQuery    string
	Limit          int32
	Offset         int32
	Actor          CbtQuestionActor
}

type ExportCbtQuestionsCSVResult struct {
	Filename string
	Content  []byte
	Count    int
}

type BulkCbtQuestionWorkflowInput struct {
	QuestionIDs []pgtype.UUID
	Action      string
	Notes       string
	Actor       CbtQuestionActor
}

type BulkCbtQuestionWorkflowItem struct {
	QuestionID pgtype.UUID `json:"question_id"`
	OK         bool        `json:"ok"`
	Error      string      `json:"error,omitempty"`
	Status     string      `json:"status,omitempty"`
	Workflow   string      `json:"workflow_status,omitempty"`
}

type BulkCbtQuestionWorkflowResult struct {
	Action  string                        `json:"action"`
	Total   int                           `json:"total"`
	Success int                           `json:"success"`
	Failed  int                           `json:"failed"`
	Items   []BulkCbtQuestionWorkflowItem `json:"items"`
}

type CbtQuestionSummary struct {
	Counts           db.GetCbtQuestionSummaryCountsRow
	BySubject        []db.ListCbtQuestionSummaryBySubjectRow
	ByCognitiveLevel []db.ListCbtQuestionSummaryByCognitiveLevelRow
	Recent           []db.ListCbtQuestionSummaryRecentRow
}

func (s *CbtQuestion) Summary(ctx context.Context, actor CbtQuestionActor) (CbtQuestionSummary, error) {
	actor = normalizeCbtQuestionActor(actor)
	base := db.GetCbtQuestionSummaryCountsParams{IsAdmin: actor.IsAdmin(), ActorUsername: actor.Username, ActorUserID: actor.UserID}
	counts, err := s.q.GetCbtQuestionSummaryCounts(ctx, base)
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	bySubject, err := s.q.ListCbtQuestionSummaryBySubject(ctx, db.ListCbtQuestionSummaryBySubjectParams(base))
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	byCognitiveLevel, err := s.q.ListCbtQuestionSummaryByCognitiveLevel(ctx, db.ListCbtQuestionSummaryByCognitiveLevelParams(base))
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	recent, err := s.q.ListCbtQuestionSummaryRecent(ctx, db.ListCbtQuestionSummaryRecentParams(base))
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	return CbtQuestionSummary{Counts: counts, BySubject: bySubject, ByCognitiveLevel: byCognitiveLevel, Recent: recent}, nil
}

func (s *CbtQuestion) ListFiltered(ctx context.Context, in ListCbtQuestionsInput) ([]db.ListCbtQuestionsFilteredRow, int64, error) {
	actor := normalizeCbtQuestionActor(in.Actor)
	rows, err := s.q.ListCbtQuestionsFiltered(ctx, db.ListCbtQuestionsFilteredParams{
		ScopeFilter:    normalizeQuestionScope(in.QuestionScope),
		EventID:        in.EventID,
		SubjectID:      in.SubjectID,
		AuthorUsername: strings.TrimSpace(in.AuthorUsername),
		WorkflowStatus: strings.TrimSpace(in.WorkflowStatus),
		StatusFilter:   normalizeQuestionStatusFilter(in.Status),
		QuestionType:   strings.TrimSpace(in.QuestionType),
		HotsFilter:     strings.TrimSpace(in.HotsFilter),
		IsAdmin:        actor.IsAdmin(),
		ActorUsername:  actor.Username,
		ActorUserID:    actor.UserID,
		RevisionSource: normalizeRevisionSource(in.RevisionSource),
		SearchQuery:    strings.TrimSpace(in.SearchQuery),
		LimitCount:     in.Limit,
		OffsetCount:    in.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountCbtQuestionsFiltered(ctx, db.CountCbtQuestionsFilteredParams{
		ScopeFilter:    normalizeQuestionScope(in.QuestionScope),
		EventID:        in.EventID,
		SubjectID:      in.SubjectID,
		AuthorUsername: strings.TrimSpace(in.AuthorUsername),
		WorkflowStatus: strings.TrimSpace(in.WorkflowStatus),
		StatusFilter:   normalizeQuestionStatusFilter(in.Status),
		QuestionType:   strings.TrimSpace(in.QuestionType),
		HotsFilter:     strings.TrimSpace(in.HotsFilter),
		IsAdmin:        actor.IsAdmin(),
		ActorUsername:  actor.Username,
		ActorUserID:    actor.UserID,
		RevisionSource: normalizeRevisionSource(in.RevisionSource),
		SearchQuery:    strings.TrimSpace(in.SearchQuery),
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
