package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microcosm-cc/bluemonday"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var importAnswerTokenSeparators = regexp.MustCompile(`[,\|;/+\s]+`)
var importMatchingPairSeparators = regexp.MustCompile(`[;,|]+`)
var bankSoalColorStyleValue = regexp.MustCompile(`(?i)^#(?:[0-9a-f]{3}|[0-9a-f]{6}|[0-9a-f]{8})$`)
var bankSoalDataAttributeValue = regexp.MustCompile(`^[a-zA-Z0-9 _.,:;@#%+=/\-()]+$`)
var bankSoalHTMLPolicy = newBankSoalHTMLPolicy()
var bankSoalPlainTextPolicy = bluemonday.StrictPolicy()

func newBankSoalHTMLPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowStandardURLs()
	p.RequireNoReferrerOnLinks(true)
	p.SkipElementsContent("script", "style", "iframe", "object", "embed", "svg", "template")

	p.AllowElements(
		"a", "b", "blockquote", "br", "code", "div", "em", "h1", "h2", "h3", "hr",
		"i", "p", "pre", "s", "span", "strong", "sub", "sup", "u",
	)
	p.AllowLists()
	p.AllowTables()
	p.AllowImages()

	p.AllowAttrs("dir").Matching(bluemonday.Direction).Globally()
	p.AllowAttrs("lang").Matching(regexp.MustCompile(`^[a-zA-Z]{2,20}$`)).Globally()
	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).Globally()
	p.AllowAttrs("data-align", "data-color").Matching(bankSoalDataAttributeValue).OnElements("span", "div", "p")
	p.AllowAttrs("style").OnElements("span", "p", "div", "h1", "h2", "h3", "td", "th")
	p.AllowStyles("text-align").MatchingEnum("left", "right", "center", "justify").Globally()
	p.AllowStyles("color").Matching(bankSoalColorStyleValue).Globally()

	return p
}

type cbtQuestionStore interface {
	CbtQuestionSubjectExists(ctx context.Context, id pgtype.UUID) (bool, error)
	CbtQuestionEventExists(ctx context.Context, id pgtype.UUID) (bool, error)
	GenerateCbtQuestionAcademicCode(ctx context.Context, arg db.GenerateCbtQuestionAcademicCodeParams) (string, error)
	ListCbtQuestions(ctx context.Context, arg db.ListCbtQuestionsParams) ([]db.ListCbtQuestionsRow, error)
	ListCbtQuestionAuthors(ctx context.Context) ([]db.ListCbtQuestionAuthorsRow, error)
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
	GetCbtQuestionAsset(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error)
	AcquireCbtQuestionDraftDuplicateLock(ctx context.Context, fingerprint string) error
	FindRecentCbtQuestionDraftDuplicate(ctx context.Context, arg db.FindRecentCbtQuestionDraftDuplicateParams) (db.CbtQuestion, error)
	CreateCbtQuestion(ctx context.Context, arg db.CreateCbtQuestionParams) (db.CbtQuestion, error)
	UpdateCbtQuestion(ctx context.Context, arg db.UpdateCbtQuestionParams) (db.CbtQuestion, error)
	GetNextCbtQuestionVersionNumber(ctx context.Context, versionGroupID pgtype.UUID) (int32, error)
	MarkCbtQuestionVersionNotLatest(ctx context.Context, id pgtype.UUID) error
	MarkCbtQuestionVersionGroupNotLatest(ctx context.Context, versionGroupID pgtype.UUID) error
	DeleteCbtQuestion(ctx context.Context, id pgtype.UUID) error
	CreateCbtQuestionAuditLog(ctx context.Context, arg db.CreateCbtQuestionAuditLogParams) (db.CbtQuestionAuditLog, error)
	CreateBankSoalQuestionWorkflowEvent(ctx context.Context, arg db.CreateBankSoalQuestionWorkflowEventParams) (db.BankSoalQuestionWorkflowEvent, error)
	CanBankSoalUserReview(ctx context.Context, arg db.CanBankSoalUserReviewParams) (bool, error)
	CanBankSoalUserApprove(ctx context.Context, arg db.CanBankSoalUserApproveParams) (bool, error)
	ListCbtQuestionTimeline(ctx context.Context, questionID pgtype.UUID) ([]db.ListCbtQuestionTimelineRow, error)
	ListBankSoalQuestionWorkflowEvents(ctx context.Context, questionID pgtype.UUID) ([]db.ListBankSoalQuestionWorkflowEventsRow, error)
	ListCbtQuestionVersions(ctx context.Context, id pgtype.UUID) ([]db.ListCbtQuestionVersionsRow, error)
}

type cbtQuestionTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type CbtQuestion struct {
	q  cbtQuestionStore
	tx cbtQuestionTxStarter
}

func NewCbtQuestion(q *db.Queries) *CbtQuestion { return &CbtQuestion{q: q} }

func NewCbtQuestionWithPool(pool *pgxpool.Pool) *CbtQuestion {
	return &CbtQuestion{q: db.New(pool), tx: pool}
}

func (s *CbtQuestion) withMutationStore(ctx context.Context, fn func(cbtQuestionStore) (db.CbtQuestion, error)) (db.CbtQuestion, error) {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	row, err := fn(db.New(tx))
	if err != nil {
		return db.CbtQuestion{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return db.CbtQuestion{}, err
	}
	committed = true
	return row, nil
}

func (s *CbtQuestion) withMutationStoreExec(ctx context.Context, fn func(cbtQuestionStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

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
	ID                   pgtype.UUID
	EventID              pgtype.UUID
	SubjectID            pgtype.UUID
	AuthoringMode        string
	Code                 string
	QuestionText         string
	QuestionType         string
	Options              []QuestionOption
	OptionA              string
	OptionB              string
	OptionC              string
	OptionD              string
	OptionE              string
	AnswerKey            string
	Explanation          string
	Difficulty           db.CbtQuestionDifficultyEnum
	Status               db.CbtQuestionStatusEnum
	StemHTML             string
	StemLatex            string
	StimulusHTML         string
	StimulusLatex        string
	ExplanationHTML      string
	RubricHTML           string
	AcademicPhase        string
	TargetLevel          string
	CPRef                string
	TPRef                string
	KDRef                string
	IndicatorRef         string
	MaterialTopic        string
	CognitiveLevel       string
	HotsFlag             bool
	MediaAssetIDs        []string
	WorkflowStatus       string
	VersionGroupID       pgtype.UUID
	VersionNumber        int32
	SourceQuestionID     pgtype.UUID
	SupersedesQuestionID pgtype.UUID
	IsLatestVersion      bool
	VersionNote          string
	AuthorUsername       string
	ReviewerUsername     string
	ApproverUsername     string
	WriterNotes          string
	ReviewNotes          string
	Actor                CbtQuestionActor
}

type CbtQuestionActor struct {
	UserID      pgtype.UUID
	Username    string
	Roles       []string
	Permissions []string
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

func (a CbtQuestionActor) HasPermission(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, permission := range a.Permissions {
		if strings.TrimSpace(permission) == target {
			return true
		}
	}
	return false
}

func (a CbtQuestionActor) CanPublishBankSoal() bool {
	return a.IsAdmin() || a.HasPermission("bank_soal.publish")
}

func (a CbtQuestionActor) CanReadAllBankSoal() bool {
	return a.IsAdmin() || a.HasPermission("bank_soal.read_all")
}

func (a CbtQuestionActor) CanUseBankSoalInPackage() bool {
	return a.HasPermission("bank_soal.use_in_package") || a.HasPermission("asesmen.package_manage")
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

func (s *CbtQuestion) ListAuthors(ctx context.Context) ([]db.ListCbtQuestionAuthorsRow, error) {
	return s.q.ListCbtQuestionAuthors(ctx)
}

type ListCbtQuestionsInput struct {
	EventID          pgtype.UUID
	QuestionScope    string
	SubjectID        pgtype.UUID
	AuthorUsername   string
	WorkflowStatus   string
	WorkflowStatuses []string
	Status           string
	QuestionType     string
	TargetLevel      string
	Difficulty       string
	CognitiveLevel   string
	MaterialTopic    string
	MetadataFilter   string
	HotsFilter       string
	RevisionSource   string
	SearchQuery      string
	SortOrder        string
	Limit            int32
	Offset           int32
	Actor            CbtQuestionActor
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
	base := db.GetCbtQuestionSummaryCountsParams{IsAdmin: actor.CanReadAllBankSoal(), CanUseInPackage: actor.CanUseBankSoalInPackage(), ActorUsername: actor.Username, ActorUserID: actor.UserID}
	counts, err := s.q.GetCbtQuestionSummaryCounts(ctx, base)
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	bySubject, err := s.q.ListCbtQuestionSummaryBySubject(ctx, db.ListCbtQuestionSummaryBySubjectParams{
		IsAdmin:         base.IsAdmin,
		CanUseInPackage: base.CanUseInPackage,
		ActorUsername:   base.ActorUsername,
		ActorUserID:     base.ActorUserID,
	})
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	byCognitiveLevel, err := s.q.ListCbtQuestionSummaryByCognitiveLevel(ctx, db.ListCbtQuestionSummaryByCognitiveLevelParams{
		IsAdmin:         base.IsAdmin,
		CanUseInPackage: base.CanUseInPackage,
		ActorUsername:   base.ActorUsername,
		ActorUserID:     base.ActorUserID,
	})
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	recent, err := s.q.ListCbtQuestionSummaryRecent(ctx, db.ListCbtQuestionSummaryRecentParams{
		IsAdmin:         base.IsAdmin,
		CanUseInPackage: base.CanUseInPackage,
		ActorUsername:   base.ActorUsername,
		ActorUserID:     base.ActorUserID,
	})
	if err != nil {
		return CbtQuestionSummary{}, err
	}
	return CbtQuestionSummary{Counts: counts, BySubject: bySubject, ByCognitiveLevel: byCognitiveLevel, Recent: recent}, nil
}

func (s *CbtQuestion) ListFiltered(ctx context.Context, in ListCbtQuestionsInput) ([]db.ListCbtQuestionsFilteredRow, int64, error) {
	actor := normalizeCbtQuestionActor(in.Actor)
	workflowStatuses := normalizeQuestionWorkflowStatuses(in.WorkflowStatuses, in.WorkflowStatus)
	rows, err := s.q.ListCbtQuestionsFiltered(ctx, db.ListCbtQuestionsFilteredParams{
		ScopeFilter:      normalizeQuestionScope(in.QuestionScope),
		EventID:          in.EventID,
		SubjectID:        in.SubjectID,
		AuthorUsername:   strings.TrimSpace(in.AuthorUsername),
		WorkflowStatuses: workflowStatuses,
		StatusFilter:     normalizeQuestionStatusFilter(in.Status),
		QuestionType:     strings.TrimSpace(in.QuestionType),
		TargetLevel:      normalizeQuestionTargetLevelFilter(in.TargetLevel),
		DifficultyFilter: normalizeQuestionDifficultyFilter(in.Difficulty),
		CognitiveLevel:   strings.TrimSpace(in.CognitiveLevel),
		MaterialTopic:    strings.TrimSpace(in.MaterialTopic),
		MetadataFilter:   normalizeQuestionMetadataFilter(in.MetadataFilter),
		HotsFilter:       strings.TrimSpace(in.HotsFilter),
		IsAdmin:          actor.CanReadAllBankSoal(),
		CanReviewAnswer:  actor.HasPermission("bank_soal.review"),
		CanApproveAnswer: actor.HasPermission("bank_soal.approve") || actor.HasPermission("bank_soal.publish"),
		CanUseInPackage:  actor.CanUseBankSoalInPackage(),
		ActorUsername:    actor.Username,
		ActorUserID:      actor.UserID,
		RevisionSource:   normalizeRevisionSource(in.RevisionSource),
		SearchQuery:      strings.TrimSpace(in.SearchQuery),
		SortOrder:        normalizeQuestionSortOrder(in.SortOrder),
		LimitCount:       in.Limit,
		OffsetCount:      in.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountCbtQuestionsFiltered(ctx, db.CountCbtQuestionsFilteredParams{
		ScopeFilter:      normalizeQuestionScope(in.QuestionScope),
		EventID:          in.EventID,
		SubjectID:        in.SubjectID,
		AuthorUsername:   strings.TrimSpace(in.AuthorUsername),
		WorkflowStatuses: workflowStatuses,
		StatusFilter:     normalizeQuestionStatusFilter(in.Status),
		QuestionType:     strings.TrimSpace(in.QuestionType),
		TargetLevel:      normalizeQuestionTargetLevelFilter(in.TargetLevel),
		DifficultyFilter: normalizeQuestionDifficultyFilter(in.Difficulty),
		CognitiveLevel:   strings.TrimSpace(in.CognitiveLevel),
		MaterialTopic:    strings.TrimSpace(in.MaterialTopic),
		MetadataFilter:   normalizeQuestionMetadataFilter(in.MetadataFilter),
		HotsFilter:       strings.TrimSpace(in.HotsFilter),
		IsAdmin:          actor.CanReadAllBankSoal(),
		CanUseInPackage:  actor.CanUseBankSoalInPackage(),
		ActorUsername:    actor.Username,
		ActorUserID:      actor.UserID,
		RevisionSource:   normalizeRevisionSource(in.RevisionSource),
		SearchQuery:      strings.TrimSpace(in.SearchQuery),
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
