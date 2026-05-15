package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type profileChangeRequestStore interface {
	GetOwnedEmployeeOfficialProfile(ctx context.Context, id pgtype.UUID) (db.GetOwnedEmployeeOfficialProfileRow, error)
	GetOwnedStudentOfficialProfile(ctx context.Context, id pgtype.UUID) (db.GetOwnedStudentOfficialProfileRow, error)
	GetOwnedParentOfficialProfile(ctx context.Context, id pgtype.UUID) (db.GetOwnedParentOfficialProfileRow, error)
	GetParentOwnedChildOfficialProfile(ctx context.Context, arg db.GetParentOwnedChildOfficialProfileParams) (db.GetParentOwnedChildOfficialProfileRow, error)
	CreateProfileChangeRequest(ctx context.Context, arg db.CreateProfileChangeRequestParams) (db.ProfileChangeRequest, error)
	ListOwnProfileChangeRequests(ctx context.Context, requesterUserID pgtype.UUID) ([]db.ListOwnProfileChangeRequestsRow, error)
	ListProfileChangeRequests(ctx context.Context, arg db.ListProfileChangeRequestsParams) ([]db.ListProfileChangeRequestsRow, error)
	CountProfileChangeRequests(ctx context.Context, arg db.CountProfileChangeRequestsParams) (int32, error)
	GetProfileChangeRequestForUpdate(ctx context.Context, id pgtype.UUID) (db.ProfileChangeRequest, error)
	CancelOwnProfileChangeRequest(ctx context.Context, arg db.CancelOwnProfileChangeRequestParams) (db.ProfileChangeRequest, error)
	ReviewProfileChangeRequest(ctx context.Context, arg db.ReviewProfileChangeRequestParams) (db.ProfileChangeRequest, error)
	UpdateEmployeeOfficialName(ctx context.Context, arg db.UpdateEmployeeOfficialNameParams) (int64, error)
	UpdateEmployeeOfficialBirthdate(ctx context.Context, arg db.UpdateEmployeeOfficialBirthdateParams) (int64, error)
	UpdateStudentOfficialName(ctx context.Context, arg db.UpdateStudentOfficialNameParams) (int64, error)
	UpdateStudentOfficialBirthdate(ctx context.Context, arg db.UpdateStudentOfficialBirthdateParams) (int64, error)
	UpdateStudentOfficialParentName(ctx context.Context, arg db.UpdateStudentOfficialParentNameParams) (int64, error)
	UpdateStudentOfficialPhone(ctx context.Context, arg db.UpdateStudentOfficialPhoneParams) (int64, error)
	UpdateStudentOfficialAddress(ctx context.Context, arg db.UpdateStudentOfficialAddressParams) (int64, error)
	UpdateParentOfficialName(ctx context.Context, arg db.UpdateParentOfficialNameParams) (int64, error)
	UpdateParentOfficialPhone(ctx context.Context, arg db.UpdateParentOfficialPhoneParams) (int64, error)
	UpdateParentOfficialAddress(ctx context.Context, arg db.UpdateParentOfficialAddressParams) (int64, error)
	UpdateParentOfficialOccupation(ctx context.Context, arg db.UpdateParentOfficialOccupationParams) (int64, error)
	UpdateParentOfficialNik(ctx context.Context, arg db.UpdateParentOfficialNikParams) (int64, error)
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type profileChangeRequestTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type ProfileChangeRequest struct {
	q  profileChangeRequestStore
	tx profileChangeRequestTxStarter
}

type CreateProfileChangeRequestInput struct {
	ProfileType     string
	TargetStudentID pgtype.UUID
	FieldKey        string
	RequestedValue  string
	Reason          string
}

type ReviewProfileChangeRequestInput struct {
	Status              string
	ReviewNote          string
	ReviewerRoles       []string
	ReviewerPermissions []string
}

type ProfileChangeRequestListFilter struct {
	Status      string
	ProfileType string
	FieldKey    string
	Search      string
}

type ProfileChangeRequestListItem struct {
	ID                   pgtype.UUID
	RequesterUserID      pgtype.UUID
	ProfileType          string
	TargetEmployeeID     pgtype.UUID
	TargetStudentID      pgtype.UUID
	TargetParentID       pgtype.UUID
	FieldKey             string
	CurrentValue         string
	RequestedValue       string
	Reason               string
	Status               db.ProfileChangeRequestStatus
	ReviewerUserID       pgtype.UUID
	ReviewNote           string
	ReviewedAt           pgtype.Timestamptz
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	RequesterUsername    string
	RequesterDisplayName string
	ReviewerUsername     string
	ReviewerDisplayName  string
	ProfileNama          string
	FieldLabel           string
}

type ownedOfficialProfile struct {
	profileType      string
	targetEmployeeID pgtype.UUID
	targetStudentID  pgtype.UUID
	targetParentID   pgtype.UUID
	currentValues    map[string]string
}

func NewProfileChangeRequest(q *db.Queries) *ProfileChangeRequest {
	return &ProfileChangeRequest{q: q}
}

func NewProfileChangeRequestWithPool(pool *pgxpool.Pool) *ProfileChangeRequest {
	return &ProfileChangeRequest{q: db.New(pool), tx: pool}
}

func (s *ProfileChangeRequest) Create(ctx context.Context, requesterUserID pgtype.UUID, input CreateProfileChangeRequestInput) (db.ProfileChangeRequest, error) {
	reason := strings.TrimSpace(input.Reason)
	if reason == "" || len(reason) > 1000 {
		return db.ProfileChangeRequest{}, domain.ErrBadRequest
	}

	var created db.ProfileChangeRequest
	err := s.withStore(ctx, func(store profileChangeRequestStore) error {
		owned, err := s.loadOwnedOfficialProfile(ctx, store, requesterUserID, input.ProfileType, input.TargetStudentID)
		if err != nil {
			return err
		}
		field, err := findProfileChangeFieldConfig(owned.profileType, input.FieldKey, true)
		if err != nil {
			return err
		}
		requestedValue, err := normalizeOfficialChangeValue(field, input.RequestedValue)
		if err != nil {
			return err
		}
		currentValue, ok := owned.currentValues[field.FieldKey]
		if !ok {
			return domain.ErrBadRequest
		}
		if requestedValue == currentValue {
			return domain.ErrBadRequest
		}
		created, err = store.CreateProfileChangeRequest(ctx, db.CreateProfileChangeRequestParams{
			RequesterUserID:  requesterUserID,
			ProfileType:      owned.profileType,
			TargetEmployeeID: owned.targetEmployeeID,
			TargetStudentID:  owned.targetStudentID,
			TargetParentID:   owned.targetParentID,
			FieldKey:         field.FieldKey,
			CurrentValue:     currentValue,
			RequestedValue:   requestedValue,
			Reason:           reason,
		})
		if err != nil {
			return err
		}
		return auditProfileChangeRequest(ctx, store, requesterUserID, "ACCOUNT_CHANGE_REQUEST_CREATED", created, map[string]any{
			"profile_type": created.ProfileType,
			"field_key":    created.FieldKey,
			"status":       string(created.Status),
		})
	})
	return created, err
}

func (s *ProfileChangeRequest) ListOwn(ctx context.Context, requesterUserID pgtype.UUID) ([]ProfileChangeRequestListItem, error) {
	rows, err := s.q.ListOwnProfileChangeRequests(ctx, requesterUserID)
	if err != nil {
		return nil, err
	}
	items := make([]ProfileChangeRequestListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, listItemFromOwnProfileChangeRequestRow(row))
	}
	return items, nil
}

func (s *ProfileChangeRequest) ListSelfRequestableFields(ctx context.Context, requesterUserID pgtype.UUID) ([]ProfileChangeRequestField, error) {
	owned, err := s.loadOwnedOfficialProfile(ctx, s.q, requesterUserID, "", pgtype.UUID{})
	if errors.Is(err, domain.ErrForbidden) {
		return []ProfileChangeRequestField{}, nil
	}
	if err != nil {
		return nil, err
	}
	return profileChangeFieldsForOwnedProfile(owned), nil
}

func (s *ProfileChangeRequest) CancelOwn(ctx context.Context, requesterUserID, id pgtype.UUID) (db.ProfileChangeRequest, error) {
	var cancelled db.ProfileChangeRequest
	err := s.withStore(ctx, func(store profileChangeRequestStore) error {
		row, err := store.CancelOwnProfileChangeRequest(ctx, db.CancelOwnProfileChangeRequestParams{
			ID:              id,
			RequesterUserID: requesterUserID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		cancelled = row
		return auditProfileChangeRequest(ctx, store, requesterUserID, "ACCOUNT_CHANGE_REQUEST_CANCELLED", row, map[string]any{
			"profile_type": row.ProfileType,
			"field_key":    row.FieldKey,
			"status":       string(row.Status),
		})
	})
	return cancelled, err
}

func (s *ProfileChangeRequest) ListAdmin(ctx context.Context, filter ProfileChangeRequestListFilter, limit, offset int32) ([]ProfileChangeRequestListItem, error) {
	params, err := profileChangeRequestListParams(filter, limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.ListProfileChangeRequests(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]ProfileChangeRequestListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, listItemFromProfileChangeRequestRow(row))
	}
	return items, nil
}

func (s *ProfileChangeRequest) CountAdmin(ctx context.Context, filter ProfileChangeRequestListFilter) (int32, error) {
	normalized, err := normalizeProfileChangeRequestListFilter(filter)
	if err != nil {
		return 0, err
	}
	return s.q.CountProfileChangeRequests(ctx, db.CountProfileChangeRequestsParams{
		StatusFilter:      normalized.Status,
		ProfileTypeFilter: normalized.ProfileType,
		FieldKeyFilter:    normalized.FieldKey,
		Search:            normalized.Search,
	})
}

func profileChangeRequestListParams(filter ProfileChangeRequestListFilter, limit, offset int32) (db.ListProfileChangeRequestsParams, error) {
	normalized, err := normalizeProfileChangeRequestListFilter(filter)
	if err != nil {
		return db.ListProfileChangeRequestsParams{}, err
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return db.ListProfileChangeRequestsParams{
		StatusFilter:      normalized.Status,
		ProfileTypeFilter: normalized.ProfileType,
		FieldKeyFilter:    normalized.FieldKey,
		Search:            normalized.Search,
		LimitCount:        limit,
		OffsetCount:       offset,
	}, nil
}

func normalizeProfileChangeRequestListFilter(filter ProfileChangeRequestListFilter) (ProfileChangeRequestListFilter, error) {
	status := strings.TrimSpace(strings.ToLower(filter.Status))
	if status == "all" {
		status = ""
	}
	if status != "" {
		parsed, err := parseProfileChangeRequestStatus(status)
		if err != nil {
			return ProfileChangeRequestListFilter{}, err
		}
		status = string(parsed)
	}

	profileType := strings.TrimSpace(strings.ToLower(filter.ProfileType))
	if profileType == "all" {
		profileType = ""
	}
	switch profileType {
	case "", "employee", "student", "parent":
	default:
		return ProfileChangeRequestListFilter{}, domain.ErrBadRequest
	}

	fieldKey, err := normalizeProfileChangeFieldFilter(filter.FieldKey)
	if err != nil {
		return ProfileChangeRequestListFilter{}, err
	}

	search := strings.Join(strings.Fields(filter.Search), " ")
	if len(search) > 120 {
		return ProfileChangeRequestListFilter{}, domain.ErrBadRequest
	}

	return ProfileChangeRequestListFilter{
		Status:      status,
		ProfileType: profileType,
		FieldKey:    fieldKey,
		Search:      search,
	}, nil
}

func (s *ProfileChangeRequest) Review(ctx context.Context, reviewerUserID, id pgtype.UUID, input ReviewProfileChangeRequestInput) (db.ProfileChangeRequest, error) {
	status, err := parseReviewProfileChangeRequestStatus(input.Status)
	if err != nil {
		return db.ProfileChangeRequest{}, err
	}
	note := strings.TrimSpace(input.ReviewNote)
	if len(note) > 1000 {
		return db.ProfileChangeRequest{}, domain.ErrBadRequest
	}
	if status == db.ProfileChangeRequestStatusRejected && note == "" {
		return db.ProfileChangeRequest{}, domain.ErrBadRequest
	}

	var reviewed db.ProfileChangeRequest
	err = s.withStore(ctx, func(store profileChangeRequestStore) error {
		current, err := store.GetProfileChangeRequestForUpdate(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		if current.Status != db.ProfileChangeRequestStatusPending {
			return domain.ErrConflict
		}
		field, err := findProfileChangeFieldConfig(current.ProfileType, current.FieldKey, false)
		if err != nil {
			return err
		}
		if !profileChangeReviewerAllowed(field, input.ReviewerRoles, input.ReviewerPermissions) {
			return domain.ErrForbidden
		}
		if status == db.ProfileChangeRequestStatusApproved {
			if err := applyApprovedProfileChangeRequest(ctx, store, current); err != nil {
				return err
			}
		}
		reviewed, err = store.ReviewProfileChangeRequest(ctx, db.ReviewProfileChangeRequestParams{
			ID:             id,
			Status:         status,
			ReviewerUserID: reviewerUserID,
			ReviewNote:     note,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrConflict
		}
		if err != nil {
			return err
		}
		action := "ACCOUNT_CHANGE_REQUEST_REJECTED"
		if status == db.ProfileChangeRequestStatusApproved {
			action = "ACCOUNT_CHANGE_REQUEST_APPROVED"
		}
		return auditProfileChangeRequest(ctx, store, reviewerUserID, action, reviewed, map[string]any{
			"profile_type": reviewed.ProfileType,
			"field_key":    reviewed.FieldKey,
			"status":       string(reviewed.Status),
		})
	})
	return reviewed, err
}

func (s *ProfileChangeRequest) withStore(ctx context.Context, fn func(profileChangeRequestStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *ProfileChangeRequest) loadOwnedOfficialProfile(ctx context.Context, store profileChangeRequestStore, userID pgtype.UUID, requestedProfileType string, targetStudentID pgtype.UUID) (ownedOfficialProfile, error) {
	requestedProfileType = strings.TrimSpace(strings.ToLower(requestedProfileType))
	if requestedProfileType != "" && requestedProfileType != "employee" && requestedProfileType != "student" && requestedProfileType != "parent" {
		return ownedOfficialProfile{}, domain.ErrBadRequest
	}
	if targetStudentID.Valid && requestedProfileType != "" && requestedProfileType != "student" {
		return ownedOfficialProfile{}, domain.ErrBadRequest
	}
	if targetStudentID.Valid {
		row, err := store.GetParentOwnedChildOfficialProfile(ctx, db.GetParentOwnedChildOfficialProfileParams{
			RequesterUserID: userID,
			TargetStudentID: targetStudentID,
		})
		if err == nil {
			return ownedOfficialProfile{
				profileType:     "student",
				targetStudentID: row.ID,
				currentValues: map[string]string{
					"nama":          row.Nama,
					"tanggal_lahir": profileChangeDateString(row.TanggalLahir),
					"parent_name":   row.ParentName,
					"phone":         row.Phone,
					"alamat":        row.Alamat,
				},
			}, nil
		}
		return ownedOfficialProfile{}, mapOwnedProfileError(err)
	}
	if requestedProfileType == "" || requestedProfileType == "employee" {
		row, err := store.GetOwnedEmployeeOfficialProfile(ctx, userID)
		if err == nil {
			return ownedOfficialProfile{
				profileType:      "employee",
				targetEmployeeID: row.ID,
				currentValues: map[string]string{
					"nama":          row.Nama,
					"tanggal_lahir": profileChangeDateString(row.TanggalLahir),
				},
			}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) || requestedProfileType == "employee" {
			return ownedOfficialProfile{}, mapOwnedProfileError(err)
		}
	}
	if requestedProfileType == "" || requestedProfileType == "student" {
		row, err := store.GetOwnedStudentOfficialProfile(ctx, userID)
		if err == nil {
			return ownedOfficialProfile{
				profileType:     "student",
				targetStudentID: row.ID,
				currentValues: map[string]string{
					"nama":          row.Nama,
					"tanggal_lahir": profileChangeDateString(row.TanggalLahir),
					"parent_name":   row.ParentName,
					"phone":         row.Phone,
					"alamat":        row.Alamat,
				},
			}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) || requestedProfileType == "student" {
			return ownedOfficialProfile{}, mapOwnedProfileError(err)
		}
	}
	if requestedProfileType == "" || requestedProfileType == "parent" {
		row, err := store.GetOwnedParentOfficialProfile(ctx, userID)
		if err == nil {
			return ownedOfficialProfile{
				profileType:    "parent",
				targetParentID: row.ID,
				currentValues: map[string]string{
					"nama":       row.Nama,
					"phone":      row.Phone,
					"address":    row.Address,
					"occupation": row.Occupation,
					"nik":        row.Nik,
				},
			}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) || requestedProfileType == "parent" {
			return ownedOfficialProfile{}, mapOwnedProfileError(err)
		}
	}
	return ownedOfficialProfile{}, domain.ErrForbidden
}

func mapOwnedProfileError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	return err
}

func parseProfileChangeRequestStatus(status string) (db.ProfileChangeRequestStatus, error) {
	switch db.ProfileChangeRequestStatus(strings.TrimSpace(strings.ToLower(status))) {
	case db.ProfileChangeRequestStatusPending:
		return db.ProfileChangeRequestStatusPending, nil
	case db.ProfileChangeRequestStatusApproved:
		return db.ProfileChangeRequestStatusApproved, nil
	case db.ProfileChangeRequestStatusRejected:
		return db.ProfileChangeRequestStatusRejected, nil
	case db.ProfileChangeRequestStatusCancelled:
		return db.ProfileChangeRequestStatusCancelled, nil
	default:
		return "", domain.ErrBadRequest
	}
}

func parseReviewProfileChangeRequestStatus(status string) (db.ProfileChangeRequestStatus, error) {
	parsed, err := parseProfileChangeRequestStatus(status)
	if err != nil {
		return "", err
	}
	if parsed != db.ProfileChangeRequestStatusApproved && parsed != db.ProfileChangeRequestStatusRejected {
		return "", domain.ErrBadRequest
	}
	return parsed, nil
}

func applyApprovedProfileChangeRequest(ctx context.Context, store profileChangeRequestStore, row db.ProfileChangeRequest) error {
	field, err := findProfileChangeFieldConfig(row.ProfileType, row.FieldKey, false)
	if err != nil {
		return err
	}
	value, err := normalizeOfficialChangeValue(field, row.RequestedValue)
	if err != nil {
		return err
	}
	var affected int64
	switch row.ProfileType {
	case "employee":
		if !row.TargetEmployeeID.Valid {
			return domain.ErrBadRequest
		}
		switch field.FieldKey {
		case "nama":
			affected, err = store.UpdateEmployeeOfficialName(ctx, db.UpdateEmployeeOfficialNameParams{ID: row.TargetEmployeeID, Nama: value})
		case "tanggal_lahir":
			affected, err = store.UpdateEmployeeOfficialBirthdate(ctx, db.UpdateEmployeeOfficialBirthdateParams{ID: row.TargetEmployeeID, TanggalLahir: mustDate(value)})
		}
	case "student":
		if !row.TargetStudentID.Valid {
			return domain.ErrBadRequest
		}
		switch field.FieldKey {
		case "nama":
			affected, err = store.UpdateStudentOfficialName(ctx, db.UpdateStudentOfficialNameParams{ID: row.TargetStudentID, Nama: value})
		case "tanggal_lahir":
			affected, err = store.UpdateStudentOfficialBirthdate(ctx, db.UpdateStudentOfficialBirthdateParams{ID: row.TargetStudentID, TanggalLahir: mustDate(value)})
		case "parent_name":
			affected, err = store.UpdateStudentOfficialParentName(ctx, db.UpdateStudentOfficialParentNameParams{ID: row.TargetStudentID, ParentName: value})
		case "phone":
			affected, err = store.UpdateStudentOfficialPhone(ctx, db.UpdateStudentOfficialPhoneParams{ID: row.TargetStudentID, Phone: value})
		case "alamat":
			affected, err = store.UpdateStudentOfficialAddress(ctx, db.UpdateStudentOfficialAddressParams{ID: row.TargetStudentID, Alamat: value})
		}
	case "parent":
		if !row.TargetParentID.Valid {
			return domain.ErrBadRequest
		}
		switch field.FieldKey {
		case "nama":
			affected, err = store.UpdateParentOfficialName(ctx, db.UpdateParentOfficialNameParams{ID: row.TargetParentID, Nama: value})
		case "phone":
			affected, err = store.UpdateParentOfficialPhone(ctx, db.UpdateParentOfficialPhoneParams{ID: row.TargetParentID, Phone: value})
		case "address":
			affected, err = store.UpdateParentOfficialAddress(ctx, db.UpdateParentOfficialAddressParams{ID: row.TargetParentID, Address: value})
		case "occupation":
			affected, err = store.UpdateParentOfficialOccupation(ctx, db.UpdateParentOfficialOccupationParams{ID: row.TargetParentID, Occupation: value})
		case "nik":
			affected, err = store.UpdateParentOfficialNik(ctx, db.UpdateParentOfficialNikParams{ID: row.TargetParentID, Nik: value})
		}
	default:
		return domain.ErrBadRequest
	}
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func mustDate(value string) pgtype.Date {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: parsed, Valid: true}
}

func profileChangeDateString(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func auditProfileChangeRequest(ctx context.Context, store profileChangeRequestStore, actorID pgtype.UUID, action string, row db.ProfileChangeRequest, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["request_id"] = uuidEntityID(row.ID)
	metadata["requester_user_id"] = uuidEntityID(row.RequesterUserID)
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = store.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     actorID,
		Action:     action,
		EntityType: "profile_change_request",
		EntityID:   uuidEntityID(row.ID),
		Metadata:   payload,
	})
	return err
}

func listItemFromOwnProfileChangeRequestRow(row db.ListOwnProfileChangeRequestsRow) ProfileChangeRequestListItem {
	reviewer := ""
	if row.ReviewerUsername.Valid {
		reviewer = row.ReviewerUsername.String
	}
	return ProfileChangeRequestListItem{
		ID:                   row.ID,
		RequesterUserID:      row.RequesterUserID,
		ProfileType:          row.ProfileType,
		TargetEmployeeID:     row.TargetEmployeeID,
		TargetStudentID:      row.TargetStudentID,
		TargetParentID:       row.TargetParentID,
		FieldKey:             row.FieldKey,
		CurrentValue:         row.CurrentValue,
		RequestedValue:       row.RequestedValue,
		Reason:               row.Reason,
		Status:               row.Status,
		ReviewerUserID:       row.ReviewerUserID,
		ReviewNote:           row.ReviewNote,
		ReviewedAt:           row.ReviewedAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		RequesterUsername:    row.RequesterUsername,
		RequesterDisplayName: row.RequesterDisplayName,
		ReviewerUsername:     reviewer,
		ReviewerDisplayName:  row.ReviewerDisplayName,
		ProfileNama:          row.ProfileNama,
		FieldLabel:           profileChangeFieldLabel(row.ProfileType, row.FieldKey),
	}
}

func listItemFromProfileChangeRequestRow(row db.ListProfileChangeRequestsRow) ProfileChangeRequestListItem {
	reviewer := ""
	if row.ReviewerUsername.Valid {
		reviewer = row.ReviewerUsername.String
	}
	return ProfileChangeRequestListItem{
		ID:                   row.ID,
		RequesterUserID:      row.RequesterUserID,
		ProfileType:          row.ProfileType,
		TargetEmployeeID:     row.TargetEmployeeID,
		TargetStudentID:      row.TargetStudentID,
		TargetParentID:       row.TargetParentID,
		FieldKey:             row.FieldKey,
		CurrentValue:         row.CurrentValue,
		RequestedValue:       row.RequestedValue,
		Reason:               row.Reason,
		Status:               row.Status,
		ReviewerUserID:       row.ReviewerUserID,
		ReviewNote:           row.ReviewNote,
		ReviewedAt:           row.ReviewedAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		RequesterUsername:    row.RequesterUsername,
		RequesterDisplayName: row.RequesterDisplayName,
		ReviewerUsername:     reviewer,
		ReviewerDisplayName:  row.ReviewerDisplayName,
		ProfileNama:          row.ProfileNama,
		FieldLabel:           profileChangeFieldLabel(row.ProfileType, row.FieldKey),
	}
}
