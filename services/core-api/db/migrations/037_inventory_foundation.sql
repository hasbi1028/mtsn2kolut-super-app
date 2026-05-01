CREATE TABLE inventory_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode          TEXT        NOT NULL UNIQUE,
    nama          TEXT        NOT NULL,
    kategori      TEXT        NOT NULL DEFAULT 'umum',
    lokasi        TEXT        NOT NULL DEFAULT '',
    kondisi       TEXT        NOT NULL DEFAULT 'baik',
    satuan        TEXT        NOT NULL DEFAULT 'unit',
    jumlah_total  INTEGER     NOT NULL DEFAULT 1,
    jumlah_baik   INTEGER     NOT NULL DEFAULT 1,
    min_stock     INTEGER     NOT NULL DEFAULT 0,
    catatan       TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inventory_kondisi CHECK (kondisi IN ('baik', 'perlu-perawatan', 'rusak')),
    CONSTRAINT chk_inventory_total CHECK (jumlah_total >= 1),
    CONSTRAINT chk_inventory_baik CHECK (jumlah_baik >= 0 AND jumlah_baik <= jumlah_total),
    CONSTRAINT chk_inventory_min_stock CHECK (min_stock >= 0)
);

CREATE INDEX idx_inventory_items_kategori ON inventory_items(kategori);
CREATE INDEX idx_inventory_items_kondisi  ON inventory_items(kondisi);
CREATE INDEX idx_inventory_items_lokasi   ON inventory_items(lokasi);
