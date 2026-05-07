-- Harden default Bank Soal teacher permissions before release.
-- Guru keeps authoring/read access, but privileged review/import/settings/publish/delete
-- must be granted explicitly by admin/RBAC assignment.
DELETE FROM rbac_role_permissions rp
USING rbac_roles r, rbac_permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.code = 'guru'
  AND p.code IN (
    'bank_soal.review',
    'bank_soal.import',
    'bank_soal.settings',
    'bank_soal.publish',
    'bank_soal.delete'
  );
