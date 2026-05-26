-- Allow backend-owned CBT portal QR+PIN starts to record their client type.
ALTER TABLE cbt_exam_participants
  DROP CONSTRAINT IF EXISTS chk_cbt_exam_participants_client_type;

ALTER TABLE cbt_exam_participants
  ADD CONSTRAINT chk_cbt_exam_participants_client_type
  CHECK (client_type = ANY (ARRAY['unknown'::text, 'android'::text, 'windows'::text, 'web_fallback'::text, 'cbt_portal'::text]));
