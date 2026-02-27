package okan

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// DiskInfo はディスクの使用状況を表す
type DiskInfo struct {
	Path       string
	TotalBytes uint64
	FreeBytes  uint64
	UsedBytes  uint64
	UsedPct    int
}

// MemInfo はメモリの使用状況を表す
type MemInfo struct {
	TotalKB     uint64
	AvailableKB uint64
	UsedKB      uint64
	UsedPct     int
}

// LoadInfo はロードアベレージを表す
type LoadInfo struct {
	Load1   float64
	Load5   float64
	Load15  float64
	CPUNum  int
	Ratio1  float64 // Load1 / CPUNum
}

// GetDiskInfo は指定パスのディスク情報を返す
func GetDiskInfo(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("ディスク情報が取れへんかった: %w", err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	pct := 0
	if total > 0 {
		pct = int(used * 100 / total)
	}

	return &DiskInfo{
		Path:       path,
		TotalBytes: total,
		FreeBytes:  free,
		UsedBytes:  used,
		UsedPct:    pct,
	}, nil
}

// GetMemInfo は /proc/meminfo からメモリ情報を返す（Linux 専用）
func GetMemInfo() (*MemInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("メモリ情報が読めへんかった: %w", err)
	}
	defer f.Close()

	info := &MemInfo{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			info.TotalKB = val
		case "MemAvailable:":
			info.AvailableKB = val
		}
	}

	info.UsedKB = info.TotalKB - info.AvailableKB
	if info.TotalKB > 0 {
		info.UsedPct = int(info.UsedKB * 100 / info.TotalKB)
	}
	return info, nil
}

// GetLoadInfo は /proc/loadavg からロードアベレージを返す（Linux 専用）
func GetLoadInfo() (*LoadInfo, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, fmt.Errorf("ロード情報が読めへんかった: %w", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return nil, fmt.Errorf("loadavg のフォーマットがおかしい")
	}

	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	l15, _ := strconv.ParseFloat(fields[2], 64)
	cpus := runtime.NumCPU()

	return &LoadInfo{
		Load1:  l1,
		Load5:  l5,
		Load15: l15,
		CPUNum: cpus,
		Ratio1: l1 / float64(cpus),
	}, nil
}

// FormatBytes はバイト数を人間が読みやすい形式に変換する
func FormatBytes(b uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// progressBar は使用率を表すASCIIプログレスバーを返す
func progressBar(pct, width int) string {
	filled := pct * width / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	// 色分け: 緑→黄→赤
	color := "\033[32m"
	if pct >= 90 {
		color = "\033[31m"
	} else if pct >= 75 {
		color = "\033[33m"
	}
	return fmt.Sprintf("%s[%s]\033[0m %d%%", color, bar, pct)
}
