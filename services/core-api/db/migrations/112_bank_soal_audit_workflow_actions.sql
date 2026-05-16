-- Hotfix: keep legacy CBT question audit log compatible with Bank Soal workflow actions.
-- The runtime writes workflow action names to cbt_question_audit_logs before
-- inserting the richer bank_soal_question_workflow_events row. Production still
-- had the legacy audit action constraint, so submit_for_review/request_revision/
-- mark_reviewed/archive failed at the audit boundary.

DO $$
BEGIN
  IF to_regclass('public.cbt_question_audit_logs') IS NOT NULL THEN
    ALTER TABLE public.cbt_question_audit_logs
      DROP CONSTRAINT IF EXISTS chk_cbt_question_audit_action;

    ALTER TABLE public.cbt_question_audit_logs
      ADD CONSTRAINT chk_cbt_question_audit_action
      CHECK (
        action = ANY (ARRAY[
          'create'::text,
          'update'::text,
          'import'::text,
          'submit_review'::text,
          'submit_for_review'::text,
          'approve'::text,
          'reject'::text,
          'publish'::text,
          'duplicate'::text,
          'revision'::text,
          'request_revision'::text,
          'mark_reviewed'::text,
          'delete'::text,
          'archive'::text
        ])
      );
  END IF;
END $$;
