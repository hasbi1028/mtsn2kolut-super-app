package service

import (
	"strings"
	"testing"
)

func TestCollectBankSoalImageAssetIDsRejectsBrokenLocalAndExternalSources(t *testing.T) {
	for name, input := range map[string]string{
		"quill_broken_placeholder": `<p><img src="//:0">Pilih gambar</p>`,
		"browser_blob":             `<p><img src="blob:http://localhost/abc"></p>`,
		"inline_base64":            `<p><img src="data:image/png;base64,AAAA"></p>`,
		"local_file":               `<p><img src="file:///tmp/a.png"></p>`,
		"external_http":            `<p><img src="https://example.com/a.png"></p>`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := collectBankSoalImageAssetIDs(input)
			if err == nil {
				t.Fatalf("collectBankSoalImageAssetIDs(%s) returned nil error", name)
			}
			if !strings.Contains(err.Error(), "server") {
				t.Fatalf("error = %q, want operator-facing server storage message", err.Error())
			}
		})
	}
}

func TestCollectBankSoalImageAssetIDsExtractsCanonicalAssetIDs(t *testing.T) {
	input := `<p><img src="/api/bank-soal/assets/11111111-1111-1111-1111-111111111111/file"></p><p><img src="/api/cbt/assets/22222222-2222-2222-2222-222222222222/file"></p><p><img src="/api/bank-soal/assets/11111111-1111-1111-1111-111111111111/file?size=sm"></p>`
	ids, err := collectBankSoalImageAssetIDs(input)
	if err != nil {
		t.Fatalf("collectBankSoalImageAssetIDs returned error: %v", err)
	}
	want := []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %#v, want %#v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %#v, want %#v", ids, want)
		}
	}
}
