package service

import (
	"strings"
	"time"

	"mtsn2kolut-super-app/backend/internal/domain"
)

const (
	ProfileChangesReviewPermission = "profile_changes.review"

	profileChangeValueText = "text"
	profileChangeValueDate = "date"
)

type ProfileChangeRequestField struct {
	ProfileType        string
	FieldKey           string
	Label              string
	ValueType          string
	SelfRequestable    bool
	ReviewerPermission string
	IsActive           bool
}

type profileChangeFieldConfig struct {
	ProfileType        string
	FieldKey           string
	Label              string
	ValueType          string
	Aliases            []string
	SelfRequestable    bool
	ReviewerPermission string
	IsActive           bool
}

var profileChangeFieldCatalog = []profileChangeFieldConfig{
	{
		ProfileType:        "employee",
		FieldKey:           "nama",
		Label:              "Nama resmi",
		ValueType:          profileChangeValueText,
		Aliases:            []string{"name"},
		SelfRequestable:    true,
		ReviewerPermission: ProfileChangesReviewPermission,
		IsActive:           true,
	},
	{
		ProfileType:        "employee",
		FieldKey:           "tanggal_lahir",
		Label:              "Tanggal lahir",
		ValueType:          profileChangeValueDate,
		Aliases:            []string{"birthdate", "birth_date"},
		SelfRequestable:    true,
		ReviewerPermission: ProfileChangesReviewPermission,
		IsActive:           true,
	},
	{
		ProfileType:        "student",
		FieldKey:           "nama",
		Label:              "Nama resmi",
		ValueType:          profileChangeValueText,
		Aliases:            []string{"name"},
		SelfRequestable:    true,
		ReviewerPermission: ProfileChangesReviewPermission,
		IsActive:           true,
	},
	{
		ProfileType:        "student",
		FieldKey:           "tanggal_lahir",
		Label:              "Tanggal lahir",
		ValueType:          profileChangeValueDate,
		Aliases:            []string{"birthdate", "birth_date"},
		SelfRequestable:    true,
		ReviewerPermission: ProfileChangesReviewPermission,
		IsActive:           true,
	},
	{
		ProfileType:        "student",
		FieldKey:           "parent_name",
		Label:              "Nama orang tua/wali",
		ValueType:          profileChangeValueText,
		Aliases:            []string{"parent_nama", "nama_orang_tua", "wali_nama"},
		SelfRequestable:    true,
		ReviewerPermission: ProfileChangesReviewPermission,
		IsActive:           true,
	},
	{
		ProfileType:        "parent",
		FieldKey:           "nama",
		Label:              "Nama resmi",
		ValueType:          profileChangeValueText,
		Aliases:            []string{"name"},
		SelfRequestable:    true,
		ReviewerPermission: ProfileChangesReviewPermission,
		IsActive:           true,
	},
}

func profileChangeRequestFieldFromConfig(config profileChangeFieldConfig) ProfileChangeRequestField {
	return ProfileChangeRequestField{
		ProfileType:        config.ProfileType,
		FieldKey:           config.FieldKey,
		Label:              config.Label,
		ValueType:          config.ValueType,
		SelfRequestable:    config.SelfRequestable,
		ReviewerPermission: config.ReviewerPermission,
		IsActive:           config.IsActive,
	}
}

func profileChangeFieldsForOwnedProfile(owned ownedOfficialProfile) []ProfileChangeRequestField {
	fields := make([]ProfileChangeRequestField, 0, len(profileChangeFieldCatalog))
	for _, config := range profileChangeFieldCatalog {
		if config.ProfileType != owned.profileType || !config.IsActive || !config.SelfRequestable {
			continue
		}
		if _, ok := owned.currentValues[config.FieldKey]; !ok {
			continue
		}
		fields = append(fields, profileChangeRequestFieldFromConfig(config))
	}
	return fields
}

func findProfileChangeFieldConfig(profileType, fieldKey string, selfRequestOnly bool) (profileChangeFieldConfig, error) {
	profileType = strings.TrimSpace(strings.ToLower(profileType))
	fieldKey = normalizeProfileChangeFieldKey(fieldKey)
	if profileType == "" || fieldKey == "" {
		return profileChangeFieldConfig{}, domain.ErrBadRequest
	}
	for _, config := range profileChangeFieldCatalog {
		if config.ProfileType != profileType || !config.IsActive {
			continue
		}
		if selfRequestOnly && !config.SelfRequestable {
			continue
		}
		if config.matchesFieldKey(fieldKey) {
			return config, nil
		}
	}
	return profileChangeFieldConfig{}, domain.ErrBadRequest
}

func normalizeProfileChangeFieldFilter(fieldKey string) (string, error) {
	fieldKey = normalizeProfileChangeFieldKey(fieldKey)
	if fieldKey == "" || fieldKey == "all" {
		return "", nil
	}
	for _, config := range profileChangeFieldCatalog {
		if !config.IsActive {
			continue
		}
		if config.matchesFieldKey(fieldKey) {
			return config.FieldKey, nil
		}
	}
	return "", domain.ErrBadRequest
}

func profileChangeFieldLabel(profileType, fieldKey string) string {
	profileType = strings.TrimSpace(strings.ToLower(profileType))
	fieldKey = normalizeProfileChangeFieldKey(fieldKey)
	for _, config := range profileChangeFieldCatalog {
		if !config.IsActive {
			continue
		}
		if profileType != "" && config.ProfileType != profileType {
			continue
		}
		if config.matchesFieldKey(fieldKey) {
			return config.Label
		}
	}
	return fieldKey
}

func normalizeOfficialChangeField(profileType, fieldKey string) (string, error) {
	config, err := findProfileChangeFieldConfig(profileType, fieldKey, true)
	if err != nil {
		return "", err
	}
	return config.FieldKey, nil
}

func normalizeOfficialChangeValue(config profileChangeFieldConfig, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domain.ErrBadRequest
	}
	switch config.ValueType {
	case profileChangeValueDate:
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return "", domain.ErrBadRequest
		}
		return parsed.Format("2006-01-02"), nil
	case profileChangeValueText:
		if len(value) > 200 {
			return "", domain.ErrBadRequest
		}
		return value, nil
	default:
		return "", domain.ErrBadRequest
	}
}

func profileChangeReviewerAllowed(config profileChangeFieldConfig, roles, permissions []string) bool {
	required := strings.TrimSpace(config.ReviewerPermission)
	if required == "" {
		return true
	}
	for _, role := range roles {
		if strings.TrimSpace(role) == "admin" {
			return true
		}
	}
	for _, permission := range permissions {
		if strings.TrimSpace(permission) == required {
			return true
		}
	}
	return false
}

func (config profileChangeFieldConfig) matchesFieldKey(fieldKey string) bool {
	if config.FieldKey == fieldKey {
		return true
	}
	for _, alias := range config.Aliases {
		if normalizeProfileChangeFieldKey(alias) == fieldKey {
			return true
		}
	}
	return false
}

func normalizeProfileChangeFieldKey(fieldKey string) string {
	fieldKey = strings.TrimSpace(strings.ToLower(fieldKey))
	return strings.ReplaceAll(fieldKey, "-", "_")
}
