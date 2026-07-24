package service

import "github.com/jackc/pgx/v5/pgtype"

// Types used by pre-existing handler tests that haven't been implemented yet.
// These are placeholder definitions to make test compilation succeed.

type UploadArchiveDocumentInput struct {
	Title     string `json:"title"`
	Category  string `json:"category"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
}

type StudentAccountGenerationResult struct {
	StudentName string `json:"student_name"`
	NIS         string `json:"nis"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Error       string `json:"error,omitempty"`
	Total       int32                          `json:"total"`
	Ready       int32                          `json:"ready"`
	Created     int32                          `json:"created"`
	Candidates  []StudentAccountGenerationCandidate `json:"candidates"`
}

type StudentAccountGenerationCandidate struct {
	StudentID         string `json:"student_id"`
	GeneratedUsername string `json:"generated_username"`
	Status            string `json:"status"`
	Error             string `json:"error,omitempty"`
}

type ParentAccountGenerationResult struct {
	ParentName string `json:"parent_name"`
	Phone      string `json:"phone"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Error      string `json:"error,omitempty"`
}

type DocumentCycleGenerateResult struct {
	ID           pgtype.UUID `json:"id"`
	DocumentName string      `json:"document_name"`
	Status       string      `json:"status"`
	OutputPath   string      `json:"output_path,omitempty"`
	Error        string      `json:"error,omitempty"`
}

type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type StatusResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ParticipantCommand struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ListNonTestAssessmentsInput struct {
	Search    string     `json:"search"`
	Status    string     `json:"status"`
	Type      string     `json:"type"`
	ClassID   pgtype.UUID `json:"class_id"`
	Page      int32      `json:"page"`
	PageSize  int32      `json:"page_size"`
}

type SaveNonTestAssessmentInput struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	MaxScore    int32  `json:"max_score"`
}

type SyncNonTestAssessmentToGradeResult struct {
	Synced int32 `json:"synced"`
	Errors int32 `json:"errors"`
}

type SaveNonTestSubmissionInput struct {
	AssessmentID string `json:"assessment_id"`
	StudentID    string `json:"student_id"`
	Score        int32  `json:"score"`
	Notes        string `json:"notes"`
}
