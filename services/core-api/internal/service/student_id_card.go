package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	StudentIDCardStatusActive    = "active"
	StudentIDCardStatusLost      = "lost"
	StudentIDCardStatusDamaged   = "damaged"
	StudentIDCardStatusRevoked   = "revoked"
	StudentIDCardStatusReplaced  = "replaced"
	StudentIDCardStatusExpired   = "expired"
	StudentIDCardStatusSuspended = "suspended"
)

type StudentIDCard struct {
	q      *db.Queries
	pepper string
}

func NewStudentIDCard(q *db.Queries) *StudentIDCard {
	pepper := strings.TrimSpace(os.Getenv("STUDENT_ID_CARD_TOKEN_PEPPER"))
	if pepper == "" {
		pepper = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}
	return &StudentIDCard{q: q, pepper: pepper}
}

type StudentIDCardIssueResult struct {
	Card  db.StudentIDCard `json:"card"`
	Token string           `json:"qr_token"`
	URL   string           `json:"qr_url"`
}

type StudentIDCardVerifyResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Card    struct {
		ID     string `json:"id"`
		CardNo string `json:"card_no"`
		Status string `json:"status"`
	} `json:"card"`
	Student struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		NIS       string `json:"nis"`
		ClassName string `json:"class_name"`
		PhotoURL  string `json:"photo_url"`
	} `json:"student"`
}

type StudentCardPortalChallenge struct {
	ChallengeID string    `json:"challenge_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	StudentHint struct {
		Name      string `json:"name"`
		ClassName string `json:"class_name"`
	} `json:"student_hint"`
}

type StudentCardPortalCompleteResult struct {
	OK        bool   `json:"ok"`
	StudentID string `json:"student_id"`
	Message   string `json:"message"`
}

func (s *StudentIDCard) List(ctx context.Context, limit, offset int32) ([]db.ListStudentIDCardsRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.q.ListStudentIDCards(ctx, db.ListStudentIDCardsParams{LimitRows: limit, OffsetRows: offset})
}

func (s *StudentIDCard) Get(ctx context.Context, id pgtype.UUID) (db.GetStudentIDCardRow, error) {
	if !id.Valid {
		return db.GetStudentIDCardRow{}, domain.ErrBadRequest
	}
	row, err := s.q.GetStudentIDCard(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return row, domain.ErrNotFound
	}
	return row, err
}

func (s *StudentIDCard) Issue(ctx context.Context, studentID, actorID pgtype.UUID, baseURL string) (StudentIDCardIssueResult, error) {
	if !studentID.Valid || !actorID.Valid {
		return StudentIDCardIssueResult{}, domain.ErrBadRequest
	}
	if existing, err := s.q.GetActiveStudentIDCardByStudent(ctx, studentID); err == nil && existing.ID.Valid {
		return StudentIDCardIssueResult{}, domain.ErrConflict
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return StudentIDCardIssueResult{}, err
	}
	token, err := randomToken(32)
	if err != nil {
		return StudentIDCardIssueResult{}, err
	}
	cardCode, err := randomToken(8)
	if err != nil {
		return StudentIDCardIssueResult{}, err
	}
	hash := s.hashToken(token)
	cardNo := fmt.Sprintf("MTSN2K-%d-%s", time.Now().Year(), strings.ToUpper(cardCode[:10]))
	card, err := s.q.CreateStudentIDCard(ctx, db.CreateStudentIDCardParams{
		StudentID: studentID, CardNo: cardNo, TokenHash: hash, TokenHint: "", CreatedByUserID: actorID, UpdatedByUserID: actorID,
	})
	if err != nil {
		return StudentIDCardIssueResult{}, err
	}
	_, _ = s.q.InsertStudentIDCardEvent(ctx, db.InsertStudentIDCardEventParams{CardID: card.ID, StudentID: card.StudentID, EventType: "generated", ActorUserID: actorID, Source: "admin", Metadata: []byte(`{}`)})
	return StudentIDCardIssueResult{Card: card, Token: token, URL: joinCardURL(baseURL, token)}, nil
}

func (s *StudentIDCard) UpdateStatus(ctx context.Context, id, actorID pgtype.UUID, status, reason string) (db.StudentIDCard, error) {
	if !id.Valid || !actorID.Valid || !validCardStatus(status) {
		return db.StudentIDCard{}, domain.ErrBadRequest
	}
	card, err := s.q.UpdateStudentIDCardStatus(ctx, db.UpdateStudentIDCardStatusParams{ID: id, Status: status, RevokedReason: pgtype.Text{String: strings.TrimSpace(reason), Valid: strings.TrimSpace(reason) != ""}, UpdatedByUserID: actorID})
	if errors.Is(err, pgx.ErrNoRows) {
		return card, domain.ErrNotFound
	}
	if err != nil {
		return card, err
	}
	_, _ = s.q.InsertStudentIDCardEvent(ctx, db.InsertStudentIDCardEventParams{CardID: card.ID, StudentID: card.StudentID, EventType: status, ActorUserID: actorID, Source: "admin", Metadata: []byte(`{}`)})
	return card, nil
}

func (s *StudentIDCard) MarkPrinted(ctx context.Context, id, actorID pgtype.UUID) (db.StudentIDCard, error) {
	card, err := s.q.MarkStudentIDCardPrinted(ctx, db.MarkStudentIDCardPrintedParams{ID: id, UpdatedByUserID: actorID})
	if errors.Is(err, pgx.ErrNoRows) {
		return card, domain.ErrNotFound
	}
	if err == nil {
		_, _ = s.q.InsertStudentIDCardEvent(ctx, db.InsertStudentIDCardEventParams{CardID: card.ID, StudentID: card.StudentID, EventType: "printed", ActorUserID: actorID, Source: "admin", Metadata: []byte(`{}`)})
	}
	return card, err
}

func (s *StudentIDCard) Reissue(ctx context.Context, oldID, actorID pgtype.UUID, baseURL, reason string) (StudentIDCardIssueResult, error) {
	old, err := s.q.GetStudentIDCard(ctx, oldID)
	if errors.Is(err, pgx.ErrNoRows) {
		return StudentIDCardIssueResult{}, domain.ErrNotFound
	}
	if err != nil {
		return StudentIDCardIssueResult{}, err
	}
	if _, err := s.UpdateStatus(ctx, oldID, actorID, StudentIDCardStatusReplaced, reason); err != nil {
		return StudentIDCardIssueResult{}, err
	}
	return s.Issue(ctx, old.StudentID, actorID, baseURL)
}

func (s *StudentIDCard) ResolveToken(ctx context.Context, token string) (db.GetStudentIDCardByHashRow, error) {
	token = extractToken(token)
	if token == "" {
		return db.GetStudentIDCardByHashRow{}, domain.ErrBadRequest
	}
	row, err := s.q.GetStudentIDCardByHash(ctx, s.hashToken(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return row, domain.ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.Status != StudentIDCardStatusActive || !row.StudentIsActive {
		return row, domain.ErrForbidden
	}
	return row, nil
}

func (s *StudentIDCard) PublicVerify(ctx context.Context, token, ip, userAgent string) (StudentIDCardVerifyResult, error) {
	row, err := s.ResolveToken(ctx, token)
	if err != nil {
		return StudentIDCardVerifyResult{Valid: false, Message: "Kartu tidak valid atau tidak aktif"}, err
	}
	_ = s.audit(ctx, row.ID, row.StudentID, pgtype.UUID{}, "public_verify", "public", ip, userAgent)
	var res StudentIDCardVerifyResult
	res.Valid = true
	res.Message = "Kartu aktif"
	res.Card.ID = idCardUUIDString(row.ID)
	res.Card.CardNo = row.CardNo
	res.Card.Status = row.Status
	res.Student.ID = idCardUUIDString(row.StudentID)
	res.Student.Name = row.Nama
	res.Student.NIS = row.Nis
	res.Student.ClassName = idCardFirstNonEmpty(row.ClassName, row.ClassCode)
	res.Student.PhotoURL = row.PhotoUrl
	return res, nil
}

func (s *StudentIDCard) StartPortalLogin(ctx context.Context, token, ip, userAgent string) (StudentCardPortalChallenge, error) {
	row, err := s.ResolveToken(ctx, token)
	if err != nil {
		return StudentCardPortalChallenge{}, err
	}
	challenge, err := randomToken(24)
	if err != nil {
		return StudentCardPortalChallenge{}, err
	}
	attempt, err := s.q.CreateStudentCardPortalLoginAttempt(ctx, db.CreateStudentCardPortalLoginAttemptParams{CardID: row.ID, StudentID: row.StudentID, Challenge: challenge, IpAddress: ip, UserAgent: userAgent, Column6: 300})
	if err != nil {
		return StudentCardPortalChallenge{}, err
	}
	_ = s.audit(ctx, row.ID, row.StudentID, pgtype.UUID{}, "portal_login_start", "portal", ip, userAgent)
	var out StudentCardPortalChallenge
	out.ChallengeID = attempt.Challenge
	out.ExpiresAt = attempt.ExpiresAt.Time
	out.StudentHint.Name = maskName(row.Nama)
	out.StudentHint.ClassName = idCardFirstNonEmpty(row.ClassName, row.ClassCode)
	return out, nil
}

func (s *StudentIDCard) CompletePortalLogin(ctx context.Context, challenge, pin string) (StudentCardPortalCompleteResult, error) {
	attempt, err := s.q.GetStartedStudentCardPortalLoginAttempt(ctx, strings.TrimSpace(challenge))
	if errors.Is(err, pgx.ErrNoRows) {
		return StudentCardPortalCompleteResult{}, domain.ErrNotFound
	}
	if err != nil {
		return StudentCardPortalCompleteResult{}, err
	}
	card, err := s.q.GetStudentIDCard(ctx, attempt.CardID)
	if errors.Is(err, pgx.ErrNoRows) {
		_, _ = s.q.UpdateStudentCardPortalLoginAttemptStatus(ctx, db.UpdateStudentCardPortalLoginAttemptStatusParams{ID: attempt.ID, Status: "failed"})
		return StudentCardPortalCompleteResult{}, domain.ErrForbidden
	}
	if err != nil {
		return StudentCardPortalCompleteResult{}, err
	}
	if card.Status != StudentIDCardStatusActive || !card.StudentIsActive {
		_, _ = s.q.UpdateStudentCardPortalLoginAttemptStatus(ctx, db.UpdateStudentCardPortalLoginAttemptStatusParams{ID: attempt.ID, Status: "failed"})
		return StudentCardPortalCompleteResult{}, domain.ErrForbidden
	}
	cred, err := s.q.GetStudentPortalPINCredential(ctx, attempt.StudentID)
	if errors.Is(err, pgx.ErrNoRows) {
		_, _ = s.q.UpdateStudentCardPortalLoginAttemptStatus(ctx, db.UpdateStudentCardPortalLoginAttemptStatusParams{ID: attempt.ID, Status: "failed"})
		return StudentCardPortalCompleteResult{}, domain.ErrUnauthorized
	}
	if err != nil {
		return StudentCardPortalCompleteResult{}, err
	}
	if !cred.IsEnabled || cred.PinHash == "" || bcrypt.CompareHashAndPassword([]byte(cred.PinHash), []byte(pin)) != nil {
		_, _ = s.q.IncrementStudentPortalPINFailure(ctx, attempt.StudentID)
		_, _ = s.q.UpdateStudentCardPortalLoginAttemptStatus(ctx, db.UpdateStudentCardPortalLoginAttemptStatusParams{ID: attempt.ID, Status: "failed"})
		return StudentCardPortalCompleteResult{}, domain.ErrUnauthorized
	}
	_, _ = s.q.ResetStudentPortalPINFailure(ctx, attempt.StudentID)
	if _, err := s.q.UpdateStudentCardPortalLoginAttemptStatus(ctx, db.UpdateStudentCardPortalLoginAttemptStatusParams{ID: attempt.ID, Status: "completed"}); err != nil {
		return StudentCardPortalCompleteResult{}, err
	}
	return StudentCardPortalCompleteResult{OK: true, StudentID: idCardUUIDString(attempt.StudentID), Message: "PIN valid. Session portal siswa diterbitkan."}, nil
}

func (s *StudentIDCard) AttendanceScan(ctx context.Context, token, activityCode, scanType string, actorID pgtype.UUID, ip, userAgent string) (db.StudentActivityAttendanceScan, error) {
	row, err := s.ResolveToken(ctx, token)
	if err != nil {
		return db.StudentActivityAttendanceScan{}, err
	}
	activityCode = strings.TrimSpace(activityCode)
	if activityCode == "" {
		activityCode = "general"
	}
	scanType = strings.TrimSpace(scanType)
	if scanType == "" {
		scanType = "checkin"
	}
	switch scanType {
	case "checkin", "checkout", "present", "late", "out":
	default:
		return db.StudentActivityAttendanceScan{}, domain.ErrBadRequest
	}
	res, err := s.q.InsertStudentActivityAttendanceScan(ctx, db.InsertStudentActivityAttendanceScanParams{CardID: row.ID, StudentID: row.StudentID, ActivityCode: activityCode, ScanType: scanType, ScannedByUserID: actorID, Metadata: []byte(`{}`)})
	if err == nil {
		_ = s.audit(ctx, row.ID, row.StudentID, actorID, "attendance_scan", "attendance", ip, userAgent)
	}
	return res, err
}

func (s *StudentIDCard) LibraryScan(ctx context.Context, token string, actorID pgtype.UUID, ip, userAgent string) (db.GetStudentIDCardByHashRow, error) {
	row, err := s.ResolveToken(ctx, token)
	if err == nil {
		_ = s.audit(ctx, row.ID, row.StudentID, actorID, "library_scan", "library", ip, userAgent)
	}
	return row, err
}

func (s *StudentIDCard) CBTValidate(ctx context.Context, token string, actorID pgtype.UUID, ip, userAgent string) ([]db.ValidateCardForCBTRow, error) {
	row, err := s.ResolveToken(ctx, token)
	if err != nil {
		return nil, err
	}
	items, err := s.q.ValidateCardForCBT(ctx, s.hashToken(extractToken(token)))
	if err == nil {
		_ = s.audit(ctx, row.ID, row.StudentID, actorID, "cbt_validate", "cbt", ip, userAgent)
	}
	return items, err
}

func (s *StudentIDCard) Events(ctx context.Context, cardID pgtype.UUID, limit, offset int32) ([]db.StudentIDCardEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.q.ListStudentIDCardEvents(ctx, db.ListStudentIDCardEventsParams{CardID: cardID, LimitRows: limit, OffsetRows: offset})
}

func (s *StudentIDCard) AuditLogs(ctx context.Context, cardID pgtype.UUID, limit, offset int32) ([]db.StudentIDCardAuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.q.ListStudentIDCardAuditLogs(ctx, db.ListStudentIDCardAuditLogsParams{CardID: cardID, LimitRows: limit, OffsetRows: offset})
}

func (s *StudentIDCard) audit(ctx context.Context, cardID, studentID, actorID pgtype.UUID, action, target, ip, userAgent string) error {
	meta, err := json.Marshal(map[string]string{"ip": ip, "ua": userAgent})
	if err != nil {
		return err
	}
	_, err = s.q.InsertStudentIDCardAuditLog(ctx, db.InsertStudentIDCardAuditLogParams{CardID: cardID, StudentID: studentID, ActorUserID: actorID, Action: action, Target: target, Metadata: meta})
	return err
}

func (s *StudentIDCard) hashToken(token string) string {
	token = extractToken(token)
	if s.pepper != "" {
		mac := hmac.New(sha256.New, []byte(s.pepper))
		mac.Write([]byte(token))
		return hex.EncodeToString(mac.Sum(nil))
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func extractToken(v string) string {
	v = strings.TrimSpace(v)
	if i := strings.LastIndex(v, "/"); i >= 0 {
		v = v[i+1:]
	}
	if i := strings.Index(v, "?"); i >= 0 {
		v = v[:i]
	}
	return strings.TrimSpace(v)
}
func joinCardURL(base, token string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return "/s/idc/" + token
	}
	return base + "/s/idc/" + token
}
func validCardStatus(s string) bool {
	switch s {
	case StudentIDCardStatusActive, StudentIDCardStatusLost, StudentIDCardStatusDamaged, StudentIDCardStatusRevoked, StudentIDCardStatusReplaced, StudentIDCardStatusExpired, StudentIDCardStatusSuspended:
		return true
	}
	return false
}
func idCardUUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}
func idCardFirstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
func maskName(name string) string {
	r := []rune(strings.TrimSpace(name))
	if len(r) <= 3 {
		return name
	}
	return string(r[:3]) + "***"
}
