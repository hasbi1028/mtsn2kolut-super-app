CREATE TYPE website_content_kind AS ENUM ('page', 'post', 'announcement');
CREATE TYPE website_content_status AS ENUM ('draft', 'published');

CREATE TABLE website_contents (
    id              UUID                   PRIMARY KEY DEFAULT gen_random_uuid(),
    kind            website_content_kind   NOT NULL,
    title           TEXT                   NOT NULL,
    slug            TEXT                   NOT NULL,
    excerpt         TEXT                   NOT NULL DEFAULT '',
    content_html    TEXT                   NOT NULL DEFAULT '',
    cover_image_url TEXT                   NOT NULL DEFAULT '',
    status          website_content_status NOT NULL DEFAULT 'draft',
    published_at    TIMESTAMPTZ,
    created_by      TEXT                   NOT NULL DEFAULT '',
    updated_by      TEXT                   NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ            NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ            NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_website_contents_kind_slug UNIQUE (kind, slug)
);

CREATE INDEX idx_website_contents_kind_status_published
    ON website_contents(kind, status, published_at DESC, created_at DESC);

INSERT INTO website_contents (kind, title, slug, excerpt, content_html, status, published_at, created_by, updated_by)
VALUES
    ('page', 'Profil Madrasah', 'profil', 'Profil singkat MTs Negeri 2 Kolaka Utara.', '<p>Lengkapi profil madrasah, sejarah singkat, visi, misi, dan informasi penting lainnya.</p>', 'published', NOW(), 'system', 'system'),
    ('page', 'Kontak', 'kontak', 'Kontak dan alamat resmi MTs Negeri 2 Kolaka Utara.', '<p>Lengkapi alamat, nomor telepon, email, dan tautan peta sekolah di halaman ini.</p>', 'published', NOW(), 'system', 'system'),
    ('page', 'Informasi PPDB', 'ppdb-info', 'Informasi jalur pendaftaran peserta didik baru.', '<p>Lengkapi syarat, jadwal, alur, dan FAQ PPDB agar terhubung dengan form pendaftaran online.</p>', 'published', NOW(), 'system', 'system')
ON CONFLICT (kind, slug) DO NOTHING;
