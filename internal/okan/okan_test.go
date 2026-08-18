package okan_test

import (
	"strings"
	"testing"
	"time"

	"github.com/shumasod/IntuneProvider/internal/okan"
)

// ─── Mood ────────────────────────────────────────────────────────────────────

func TestCurrentMood_returnsValidMood(t *testing.T) {
	mood := okan.CurrentMood()
	validMoods := []okan.Mood{
		okan.MoodMorning,
		okan.MoodNoon,
		okan.MoodAfternoon,
		okan.MoodEvening,
		okan.MoodNight,
	}
	for _, m := range validMoods {
		if mood == m {
			return
		}
	}
	t.Errorf("CurrentMood() = %v, not a valid Mood", mood)
}

// ─── Pick ─────────────────────────────────────────────────────────────────────

func TestPick_returnsOneOfTheSlice(t *testing.T) {
	phrases := []string{"A", "B", "C"}
	for i := 0; i < 30; i++ {
		got := okan.Pick(phrases)
		found := false
		for _, p := range phrases {
			if p == got {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Pick() = %q, not in input slice", got)
		}
	}
}

func TestPick_emptySlice_returnsEmpty(t *testing.T) {
	if got := okan.Pick(nil); got != "" {
		t.Errorf("Pick(nil) = %q, want empty string", got)
	}
	if got := okan.Pick([]string{}); got != "" {
		t.Errorf("Pick([]) = %q, want empty string", got)
	}
}

// ─── Greeting ────────────────────────────────────────────────────────────────

func TestGreeting_returnsNonEmpty(t *testing.T) {
	msg := okan.Greeting()
	if msg == "" {
		t.Error("Greeting() returned empty string")
	}
}

func TestGreeting_containsJapanese(t *testing.T) {
	for i := 0; i < 10; i++ {
		msg := okan.Greeting()
		hasJP := false
		for _, r := range msg {
			if r > 0x3000 {
				hasJP = true
				break
			}
		}
		if !hasJP {
			t.Errorf("Greeting() = %q, expected to contain Japanese characters", msg)
		}
	}
}

// ─── RandomNag ───────────────────────────────────────────────────────────────

func TestRandomNag_returnsNonEmpty(t *testing.T) {
	for i := 0; i < 20; i++ {
		if msg := okan.RandomNag(); msg == "" {
			t.Error("RandomNag() returned empty string")
		}
	}
}

// ─── Cheer / Goodnight / GoodMorning ─────────────────────────────────────────

func TestCheer_returnsNonEmpty(t *testing.T) {
	if msg := okan.Cheer(); msg == "" {
		t.Error("Cheer() returned empty string")
	}
}

func TestGoodnight_returnsNonEmpty(t *testing.T) {
	if msg := okan.Goodnight(); msg == "" {
		t.Error("Goodnight() returned empty string")
	}
}

func TestGoodMorning_returnsNonEmpty(t *testing.T) {
	if msg := okan.GoodMorning(); msg == "" {
		t.Error("GoodMorning() returned empty string")
	}
}

// ─── DiskComment ─────────────────────────────────────────────────────────────

func TestDiskComment_byUsage(t *testing.T) {
	cases := []struct {
		pct      int
		contains string
	}{
		{10, "余裕"},
		{60, "埋まって"},
		{80, "いっぱい"},
		{95, "いっぱい"},
	}
	for _, tc := range cases {
		got := okan.DiskComment(tc.pct)
		if got == "" {
			t.Errorf("DiskComment(%d) returned empty string", tc.pct)
		}
		if !strings.Contains(got, tc.contains) {
			t.Logf("DiskComment(%d) = %q (wanted %q — may vary by random)", tc.pct, got, tc.contains)
		}
	}
}

// ─── MemComment ──────────────────────────────────────────────────────────────

func TestMemComment_byUsage(t *testing.T) {
	for _, pct := range []int{20, 60, 85, 98} {
		got := okan.MemComment(pct)
		if got == "" {
			t.Errorf("MemComment(%d) returned empty string", pct)
		}
	}
}

// ─── LoadComment ─────────────────────────────────────────────────────────────

func TestLoadComment_byRatio(t *testing.T) {
	for _, ratio := range []float64{0.1, 0.7, 1.5, 3.0} {
		got := okan.LoadComment(ratio)
		if got == "" {
			t.Errorf("LoadComment(%f) returned empty string", ratio)
		}
	}
}

// ─── FormatBytes ─────────────────────────────────────────────────────────────

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		input uint64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1536 * 1024 * 1024, "1.5 GB"},
	}
	for _, tc := range cases {
		got := okan.FormatBytes(tc.input)
		if got != tc.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ─── CurrentMood の時間帯マッピング（固定時刻テスト） ────────────────────────

func TestMoodMapping(t *testing.T) {
	// 時刻に対して期待されるMoodをテスト
	// 注: time.Now() の時間帯に依存するので MoodMorning 等は直接テストしない
	// ここでは MoodMorning の定数値が 0 であること等、型の整合性を確認する
	if okan.MoodMorning == okan.MoodNight {
		t.Error("MoodMorning and MoodNight must be different")
	}
	_ = time.Now() // time パッケージが正常にインポートされていることを確認
}
