package okan_test

import (
	"strings"
	"testing"

	"github.com/shumasod/IntuneProvider/internal/okan"
)

// ─── GetDiskInfo ─────────────────────────────────────────────────────────────

func TestGetDiskInfo_rootPath(t *testing.T) {
	info, err := okan.GetDiskInfo("/")
	if err != nil {
		t.Fatalf("GetDiskInfo(\"/\") error: %v", err)
	}
	if info.TotalBytes == 0 {
		t.Error("TotalBytes should be > 0")
	}
	if info.UsedPct < 0 || info.UsedPct > 100 {
		t.Errorf("UsedPct = %d, want 0-100", info.UsedPct)
	}
	if info.UsedBytes+info.FreeBytes > info.TotalBytes+1024 {
		t.Errorf("UsedBytes(%d)+FreeBytes(%d) > TotalBytes(%d)",
			info.UsedBytes, info.FreeBytes, info.TotalBytes)
	}
}

func TestGetDiskInfo_invalidPath(t *testing.T) {
	_, err := okan.GetDiskInfo("/nonexistent/path/xyz")
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}
}

// ─── GetMemInfo ──────────────────────────────────────────────────────────────

func TestGetMemInfo(t *testing.T) {
	info, err := okan.GetMemInfo()
	if err != nil {
		t.Fatalf("GetMemInfo() error: %v", err)
	}
	if info.TotalKB == 0 {
		t.Error("TotalKB should be > 0")
	}
	if info.UsedPct < 0 || info.UsedPct > 100 {
		t.Errorf("UsedPct = %d, want 0-100", info.UsedPct)
	}
	if info.AvailableKB > info.TotalKB {
		t.Errorf("AvailableKB(%d) > TotalKB(%d)", info.AvailableKB, info.TotalKB)
	}
}

// ─── GetLoadInfo ─────────────────────────────────────────────────────────────

func TestGetLoadInfo(t *testing.T) {
	info, err := okan.GetLoadInfo()
	if err != nil {
		t.Fatalf("GetLoadInfo() error: %v", err)
	}
	if info.CPUNum <= 0 {
		t.Errorf("CPUNum = %d, want > 0", info.CPUNum)
	}
	if info.Load1 < 0 {
		t.Errorf("Load1 = %f, want >= 0", info.Load1)
	}
	if info.Ratio1 < 0 {
		t.Errorf("Ratio1 = %f, want >= 0", info.Ratio1)
	}
}

// ─── FormatBytes (追加ケース) ─────────────────────────────────────────────────

func TestFormatBytes_units(t *testing.T) {
	cases := []struct {
		bytes  uint64
		suffix string
	}{
		{100, "B"},
		{2048, "KB"},
		{3 * 1024 * 1024, "MB"},
		{5 * 1024 * 1024 * 1024, "GB"},
	}
	for _, tc := range cases {
		got := okan.FormatBytes(tc.bytes)
		if !strings.HasSuffix(got, tc.suffix) {
			t.Errorf("FormatBytes(%d) = %q, want suffix %q", tc.bytes, got, tc.suffix)
		}
	}
}
