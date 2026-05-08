package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const maxUserProfileCandidateLimit int32 = 100

type UserProfileCandidateFilter struct {
	Role          string
	ClassID       string
	Query         string
	IncludeLinked bool
	Limit         int32
}

type UserProfileCandidateChild struct {
	ID        string `json:"id"`
	Nama      string `json:"nama"`
	ClassName string `json:"class_name"`
}

type UserProfileCandidate struct {
	ID             string                      `json:"id"`
	ProfileType    string                      `json:"profile_type"`
	Nama           string                      `json:"nama"`
	Identifier     string                      `json:"identifier"`
	ClassID        string                      `json:"class_id,omitempty"`
	ClassName      string                      `json:"class_name,omitempty"`
	LinkedUserID   string                      `json:"linked_user_id,omitempty"`
	LinkedUsername string                      `json:"linked_username,omitempty"`
	IsLinked       bool                        `json:"is_linked"`
	Children       []UserProfileCandidateChild `json:"children"`
}

type UserProfileCandidatesResult struct {
	Role       string                 `json:"role"`
	Profile    string                 `json:"profile"`
	ClassID    string                 `json:"class_id,omitempty"`
	Candidates []UserProfileCandidate `json:"candidates"`
}

type userProfileCandidateStore interface {
	ListEmployeeProfileCandidates(ctx context.Context, arg db.ListEmployeeProfileCandidatesParams) ([]db.ListEmployeeProfileCandidatesRow, error)
	ListStudentProfileCandidatesByClass(ctx context.Context, arg db.ListStudentProfileCandidatesByClassParams) ([]db.ListStudentProfileCandidatesByClassRow, error)
	ListParentProfileCandidatesByChildClass(ctx context.Context, arg db.ListParentProfileCandidatesByChildClassParams) ([]db.ListParentProfileCandidatesByChildClassRow, error)
}

type UserProfileCandidateService struct {
	store userProfileCandidateStore
}

func NewUserProfileCandidateService(store userProfileCandidateStore) *UserProfileCandidateService {
	return &UserProfileCandidateService{store: store}
}

func (s *UserProfileCandidateService) List(ctx context.Context, filter UserProfileCandidateFilter) (UserProfileCandidatesResult, error) {
	role := strings.ToLower(strings.TrimSpace(filter.Role))
	profile := profileTypeForRole(role)
	if profile == "" {
		return UserProfileCandidatesResult{}, domain.ErrBadRequest
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > maxUserProfileCandidateLimit {
		limit = maxUserProfileCandidateLimit
	}
	query := strings.TrimSpace(filter.Query)

	result := UserProfileCandidatesResult{Role: role, Profile: profile, ClassID: strings.TrimSpace(filter.ClassID), Candidates: []UserProfileCandidate{}}

	switch profile {
	case "employee":
		rows, err := s.store.ListEmployeeProfileCandidates(ctx, db.ListEmployeeProfileCandidatesParams{
			IncludeLinked: filter.IncludeLinked,
			Q:             query,
			LimitCount:    limit,
		})
		if err != nil {
			return UserProfileCandidatesResult{}, err
		}
		for _, row := range rows {
			result.Candidates = append(result.Candidates, UserProfileCandidate{
				ID:             pgUUIDString(row.ID),
				ProfileType:    "employee",
				Nama:           row.Nama,
				Identifier:     row.Identifier,
				LinkedUserID:   pgUUIDString(row.LinkedUserID),
				LinkedUsername: row.LinkedUsername,
				IsLinked:       row.LinkedUserID.Valid,
				Children:       []UserProfileCandidateChild{},
			})
		}
		return result, nil
	case "student":
		classID, err := parseServiceUUID(filter.ClassID)
		if err != nil {
			return UserProfileCandidatesResult{}, domain.ErrBadRequest
		}
		result.ClassID = classID.String()
		rows, err := s.store.ListStudentProfileCandidatesByClass(ctx, db.ListStudentProfileCandidatesByClassParams{
			ClassID:       classID,
			IncludeLinked: filter.IncludeLinked,
			Q:             query,
			LimitCount:    limit,
		})
		if err != nil {
			return UserProfileCandidatesResult{}, err
		}
		for _, row := range rows {
			result.Candidates = append(result.Candidates, UserProfileCandidate{
				ID:             pgUUIDString(row.ID),
				ProfileType:    "student",
				Nama:           row.Nama,
				Identifier:     row.Identifier,
				ClassID:        pgUUIDString(row.ClassID),
				ClassName:      row.ClassName,
				LinkedUserID:   pgUUIDString(row.LinkedUserID),
				LinkedUsername: row.LinkedUsername,
				IsLinked:       row.LinkedUserID.Valid,
				Children:       []UserProfileCandidateChild{},
			})
		}
		return result, nil
	case "parent":
		classID, err := parseServiceUUID(filter.ClassID)
		if err != nil {
			return UserProfileCandidatesResult{}, domain.ErrBadRequest
		}
		result.ClassID = classID.String()
		rows, err := s.store.ListParentProfileCandidatesByChildClass(ctx, db.ListParentProfileCandidatesByChildClassParams{
			ClassID:       classID,
			IncludeLinked: filter.IncludeLinked,
			Q:             query,
			LimitCount:    limit,
		})
		if err != nil {
			return UserProfileCandidatesResult{}, err
		}
		byID := map[string]int{}
		for _, row := range rows {
			id := pgUUIDString(row.ID)
			idx, ok := byID[id]
			if !ok {
				result.Candidates = append(result.Candidates, UserProfileCandidate{
					ID:             id,
					ProfileType:    "parent",
					Nama:           row.Nama,
					Identifier:     maskPhoneIdentifier(row.Identifier),
					ClassID:        pgUUIDString(row.ClassID),
					ClassName:      row.ClassName,
					LinkedUserID:   pgUUIDString(row.LinkedUserID),
					LinkedUsername: row.LinkedUsername,
					IsLinked:       row.LinkedUserID.Valid,
					Children:       []UserProfileCandidateChild{},
				})
				idx = len(result.Candidates) - 1
				byID[id] = idx
			}
			result.Candidates[idx].Children = append(result.Candidates[idx].Children, UserProfileCandidateChild{
				ID:        pgUUIDString(row.ChildID),
				Nama:      row.ChildNama,
				ClassName: row.ClassName,
			})
		}
		return result, nil
	default:
		return UserProfileCandidatesResult{}, domain.ErrBadRequest
	}
}

func profileTypeForRole(role string) string {
	switch role {
	case "guru", "staf", "kesiswaan", "employee", "pegawai":
		return "employee"
	case "siswa", "student":
		return "student"
	case "ortu", "orang_tua", "parent":
		return "parent"
	default:
		return ""
	}
}

func parseServiceUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	return id, id.Scan(strings.TrimSpace(value))
}

func maskPhoneIdentifier(value string) string {
	digits := make([]rune, 0, len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) == 0 {
		return ""
	}
	if len(digits) <= 4 {
		return "****"
	}
	return "******" + string(digits[len(digits)-4:])
}
