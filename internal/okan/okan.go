// Package okan implements the おかん (Okan) command — a Linux-style CLI tool
// that monitors your system and nags you in Kansai dialect, just like a mother would.
package okan

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const banner = `
  ／￣￣＼
 ／ ・  ・ ＼
│  ▼    ▼  │   おかん %s
│  ∧___∧  │   ─────────────────────────────
 ＼  ＿＿  ／   %s
  ＼＿＿＿／
`

var version = "dev"

// SetVersion はビルド時にバージョンを注入する
func SetVersion(v string) { version = v }

// Execute はおかんコマンドを起動する
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "okan",
	Short: "おかん — システムを監視して関西弁で説教するコマンド",
	Long: `おかん — あなたのコンピュータのお母さん

時間帯やシステムの状態に応じて、おかんが関西弁で心配してくれます。
怒られたり、心配されたり、応援してもらったり。

COMMANDS
  okan              今の気分で話しかけてくれる
  okan neru         おやすみを言う
  okan okita        おはようを言う
  okan ganbare      応援してもらう
  okan disk [path]  ディスクを心配してもらう
  okan mem          メモリを心配してもらう
  okan cpu          CPUを心配してもらう
  okan joho         システム全体を心配してもらう
  okan kogoto       ランダムに小言を言われる
  okan watch        定期的に小言を言われ続ける
  okan version      バージョン確認

TIPS
  怒られても怒らんといてね。全部愛情やで。`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		say(Greeting())
		return nil
	},
}

// say はおかんの顔と一緒にメッセージを表示する
func say(msg string) {
	// 長いメッセージは折り返し
	wrapped := wrapText(msg, 40)
	fmt.Printf(banner, version, wrapped)
}

// saySimple はアスキーアートなしで表示する（watch モード用）
func saySimple(msg string) {
	t := time.Now().Format("15:04:05")
	fmt.Printf("\033[33m[おかん %s]\033[0m %s\n", t, msg)
}

// wrapText はテキストを指定幅で折り返す（日本語対応）
func wrapText(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	var sb strings.Builder
	for i, r := range runes {
		if i > 0 && i%width == 0 {
			sb.WriteString("\n               ") // バナーの右側に合わせるインデント
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func init() {
	rootCmd.AddCommand(neruCmd)
	rootCmd.AddCommand(okitaCmd)
	rootCmd.AddCommand(ganbareCmd)
	rootCmd.AddCommand(diskCmd)
	rootCmd.AddCommand(memCmd)
	rootCmd.AddCommand(cpuCmd)
	rootCmd.AddCommand(johoCmd)
	rootCmd.AddCommand(kogotoCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "バージョン確認",
		Run: func(cmd *cobra.Command, args []string) {
			say(fmt.Sprintf("おかん version %s やよ。ちゃんと最新版使いなさいよ！", version))
		},
	})
}

// ─── neru ──────────────────────────────────────────────────────────────────

var neruCmd = &cobra.Command{
	Use:   "neru",
	Short: "おやすみを言う",
	Long:  "寝る前においで。おかんがおやすみを言うてくれるよ。",
	Run: func(cmd *cobra.Command, args []string) {
		say(Goodnight())
	},
}

// ─── okita ─────────────────────────────────────────────────────────────────

var okitaCmd = &cobra.Command{
	Use:   "okita",
	Short: "おはようを言う",
	Long:  "起きたら一番におかんに報告しなさいよ。",
	Run: func(cmd *cobra.Command, args []string) {
		say(GoodMorning())
	},
}

// ─── ganbare ───────────────────────────────────────────────────────────────

var ganbareCmd = &cobra.Command{
	Use:   "ganbare",
	Short: "応援してもらう",
	Long:  "落ち込んでる時はおかんに励ましてもらいなさい。",
	Run: func(cmd *cobra.Command, args []string) {
		say(Cheer())
	},
}

// ─── disk ──────────────────────────────────────────────────────────────────

var diskCmd = &cobra.Command{
	Use:   "disk [path]",
	Short: "ディスクを心配してもらう  (df っぽいやつ)",
	Long:  "ディスクの空き容量を見て、おかんが心配してくれるよ。",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/"
		if len(args) > 0 {
			path = args[0]
		}
		info, err := GetDiskInfo(path)
		if err != nil {
			return err
		}

		fmt.Printf("\n\033[1m%s のディスク状況\033[0m\n", path)
		fmt.Printf("  合計    : %s\n", FormatBytes(info.TotalBytes))
		fmt.Printf("  使用済  : %s\n", FormatBytes(info.UsedBytes))
		fmt.Printf("  空き    : %s\n", FormatBytes(info.FreeBytes))
		fmt.Printf("  使用率  : %s\n\n", progressBar(info.UsedPct, 20))

		say(DiskComment(info.UsedPct))
		return nil
	},
}

// ─── mem ───────────────────────────────────────────────────────────────────

var memCmd = &cobra.Command{
	Use:   "mem",
	Short: "メモリを心配してもらう  (free っぽいやつ)",
	Long:  "メモリの使用状況を見て、おかんが心配してくれるよ。",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := GetMemInfo()
		if err != nil {
			return err
		}

		fmt.Printf("\n\033[1mメモリ状況\033[0m\n")
		fmt.Printf("  合計    : %s\n", FormatBytes(info.TotalKB*1024))
		fmt.Printf("  使用済  : %s\n", FormatBytes(info.UsedKB*1024))
		fmt.Printf("  利用可  : %s\n", FormatBytes(info.AvailableKB*1024))
		fmt.Printf("  使用率  : %s\n\n", progressBar(info.UsedPct, 20))

		say(MemComment(info.UsedPct))
		return nil
	},
}

// ─── cpu ───────────────────────────────────────────────────────────────────

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "CPUを心配してもらう  (uptime っぽいやつ)",
	Long:  "ロードアベレージを見て、おかんが心配してくれるよ。",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := GetLoadInfo()
		if err != nil {
			return err
		}

		fmt.Printf("\n\033[1mCPU / ロードアベレージ\033[0m\n")
		fmt.Printf("  CPUコア数  : %d\n", info.CPUNum)
		fmt.Printf("   1分平均   : %.2f\n", info.Load1)
		fmt.Printf("   5分平均   : %.2f\n", info.Load5)
		fmt.Printf("  15分平均   : %.2f\n", info.Load15)
		fmt.Printf("  負荷率     : %.0f%%\n\n", info.Ratio1*100)

		say(LoadComment(info.Ratio1))
		return nil
	},
}

// ─── joho (情報) ───────────────────────────────────────────────────────────

var johoCmd = &cobra.Command{
	Use:   "joho",
	Short: "システム全体を心配してもらう  (top っぽいやつ)",
	Long:  "ディスク・メモリ・CPUを全部まとめて心配してもらうよ。",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("\033[1m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m")
		fmt.Println("\033[1m         おかんのシステム点検              \033[0m")
		fmt.Printf("         %s\n", time.Now().Format("2006/01/02 15:04:05"))
		fmt.Println("\033[1m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")

		// ディスク
		if disk, err := GetDiskInfo("/"); err == nil {
			fmt.Printf("💾 \033[1mディスク (/)  \033[0m%s\n", progressBar(disk.UsedPct, 20))
			fmt.Printf("   %s / %s 使用済\n\n", FormatBytes(disk.UsedBytes), FormatBytes(disk.TotalBytes))
		}

		// メモリ
		if mem, err := GetMemInfo(); err == nil {
			fmt.Printf("🧠 \033[1mメモリ        \033[0m%s\n", progressBar(mem.UsedPct, 20))
			fmt.Printf("   %s / %s 使用済\n\n", FormatBytes(mem.UsedKB*1024), FormatBytes(mem.TotalKB*1024))
		}

		// CPU
		if load, err := GetLoadInfo(); err == nil {
			pct := int(load.Ratio1 * 100)
			if pct > 100 {
				pct = 100
			}
			fmt.Printf("⚡ \033[1mCPU負荷       \033[0m%s\n", progressBar(pct, 20))
			fmt.Printf("   load: %.2f / %.2f / %.2f  (%d core)\n\n",
				load.Load1, load.Load5, load.Load15, load.CPUNum)
		}

		fmt.Println("\033[1m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m")

		// 一番しんどそうなところを優先してコメント
		msg := Greeting()
		if disk, err := GetDiskInfo("/"); err == nil && disk.UsedPct >= 75 {
			msg = DiskComment(disk.UsedPct)
		} else if mem, err := GetMemInfo(); err == nil && mem.UsedPct >= 75 {
			msg = MemComment(mem.UsedPct)
		} else if load, err := GetLoadInfo(); err == nil && load.Ratio1 >= 1.0 {
			msg = LoadComment(load.Ratio1)
		}
		say(msg)
		return nil
	},
}

// ─── kogoto (小言) ─────────────────────────────────────────────────────────

var kogotoCmd = &cobra.Command{
	Use:   "kogoto",
	Short: "ランダムに小言を言われる",
	Long:  "何もしてなくても小言を言われる。それがおかんやで。",
	Run: func(cmd *cobra.Command, args []string) {
		say(RandomNag())
	},
}

// ─── watch ─────────────────────────────────────────────────────────────────

var watchInterval int

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "定期的に小言を言われ続ける  (watch っぽいやつ)",
	Long: `指定した間隔（分）でおかんが小言を言い続けるよ。
Ctrl+C で止められるよ。止めたらおかんが寂しがるから気をつけてね。`,
	Run: func(cmd *cobra.Command, args []string) {
		interval := time.Duration(watchInterval) * time.Minute
		saySimple(fmt.Sprintf("ほな%d分おきに様子見るよ。Ctrl+Cで止められるけど、止めたら寂しいやんか。", watchInterval))

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		nags := []func() string{
			Greeting,
			RandomNag,
			func() string {
				if disk, err := GetDiskInfo("/"); err == nil {
					return DiskComment(disk.UsedPct)
				}
				return RandomNag()
			},
			func() string {
				if mem, err := GetMemInfo(); err == nil {
					return MemComment(mem.UsedPct)
				}
				return RandomNag()
			},
			Cheer,
			RandomNag,
		}

		i := 0
		for range ticker.C {
			saySimple(nags[i%len(nags)]())
			i++
		}
	},
}

func init() {
	watchCmd.Flags().IntVarP(&watchInterval, "interval", "n", 30, "小言の間隔（分）")
}
