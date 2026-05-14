-- Add branding settings permission and grant it to admin.
INSERT INTO rbac_permissions (code, module, action, description)
VALUES ('settings.branding', 'settings', 'branding', 'Mengelola logo, favicon, ikon PWA, dan identitas visual aplikasi.')
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description;

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
JOIN rbac_permissions p ON p.code = 'settings.branding'
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO app_settings (key, value)
VALUES
  ('branding.app_name', 'MTs Negeri 2 Kolaka Utara'),
  ('branding.short_name', 'MTsN 2 Kolut'),
  ('branding.tagline', 'Super App Madrasah'),
  ('branding.primary_color', '#166534'),
  ('branding.theme_color', '#166534'),
  ('branding.logo_url', '/brand/madrasah-mark.svg'),
  ('branding.mark_url', '/brand/madrasah-mark.svg'),
  ('branding.favicon_url', '/favicon.ico'),
  ('branding.apple_touch_icon_url', '/apple-touch-icon.png'),
  ('branding.pwa_icon_192_url', '/pwa-icon-192.png'),
  ('branding.pwa_icon_512_url', '/pwa-icon-512.png'),
  ('branding.formal_logo_url', '/brand/logo-kemenag.png'),
  ('branding.version', 'default')
ON CONFLICT (key) DO NOTHING;
