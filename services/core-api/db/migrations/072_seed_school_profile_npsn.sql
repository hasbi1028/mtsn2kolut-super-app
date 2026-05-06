-- Seed MTsN 2 Kolaka Utara NPSN for username generation.
-- Existing non-empty value is preserved so future NPSN corrections from settings are not overwritten.
INSERT INTO app_settings (key, value, updated_at)
VALUES ('school_profile.npsn', '40406031', NOW())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW()
WHERE btrim(app_settings.value) = '';
