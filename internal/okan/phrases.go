package okan

import (
	"math/rand"
	"time"
)

// Mood は時間帯に応じたおかんの気分を表す
type Mood int

const (
	MoodMorning   Mood = iota // 朝 (5-11)
	MoodNoon                  // 昼 (11-14)
	MoodAfternoon             // 午後 (14-18)
	MoodEvening               // 夜 (18-22)
	MoodNight                 // 深夜 (22-25/0-5)
)

// CurrentMood は現在時刻からおかんの気分を返す
func CurrentMood() Mood {
	h := time.Now().Hour()
	switch {
	case h >= 5 && h < 11:
		return MoodMorning
	case h >= 11 && h < 14:
		return MoodNoon
	case h >= 14 && h < 18:
		return MoodAfternoon
	case h >= 18 && h < 22:
		return MoodEvening
	default:
		return MoodNight
	}
}

// greetings は時間帯ごとのあいさつ・ガミガミ
var greetings = map[Mood][]string{
	MoodMorning: {
		"おはよう！ちゃんと朝ごはん食べた？食べてへんやろ、もう！",
		"おはようさん！今日も元気に頑張りなさいよ！",
		"起きたんか。顔洗ってごはん食べてから仕事しなさいよ。",
		"朝から頑張ってるやないの！でもちゃんとご飯食べなあかんよ！",
	},
	MoodNoon: {
		"お昼やで！ちゃんとご飯食べた？食べんと体壊すよ！",
		"昼休みはちゃんと休まなあかんよ。パソコンから離れなさい！",
		"お昼ご飯はちゃんと食べた？カップラーメンとかやないよね？",
		"外に出て日なたぼっこでもしてきなさいよ。ビタミンD大事やで！",
	},
	MoodAfternoon: {
		"3時のおやつは食べた？ちょっと一息つきなさいよ。",
		"お茶でも飲んで休憩しなさい。根詰めすぎやで。",
		"目が疲れてへん？20分に一回は目を休めなさいって言ったやろ！",
		"背筋伸ばして！ずっと猫背で座ってたら腰痛になるよ！",
	},
	MoodEvening: {
		"もうこんな時間やのに！夕飯食べた？",
		"今日も一日お疲れさん！ちゃんとご飯食べてゆっくりしなさいよ。",
		"残業はほどほどにしなさいよ！体が資本やで。",
		"仕事ばっかりやなくて、たまには趣味の時間も作りなさいよ。",
	},
	MoodNight: {
		"こんな夜中に何してるの！今すぐ寝なさい！",
		"もう何時だと思ってるの！絶対体壊すよ、ほんまに心配や！",
		"夜更かしし過ぎ！明日もあるんやから早く寝なさい！",
		"睡眠不足は万病の元やって知ってる？今すぐパソコン閉めなさい！",
	},
}

// diskPhrases はディスク使用率に応じたセリフ
var diskPhrases = []struct {
	threshold int
	phrases   []string
}{
	{50, []string{
		"ディスク、まだ余裕あるね。でも整理整頓はしとき！",
		"今のうちに要らんファイルは捨てときなさいよ。",
	}},
	{75, []string{
		"ディスクがだいぶ埋まってきてるよ。そろそろ整理しなさい！",
		"不要なファイルは溜めといたらあかんよ！今すぐ整理しなさい！",
	}},
	{90, []string{
		"ディスクがいっぱいになってきてるよ！早く整理しなさい！",
		"こんなにデータ溜めて！物を大事にするのはええけど整理もしなさい！",
	}},
	{101, []string{
		"ディスクがほぼいっぱいやないの！今すぐ整理しなさい！動かんくなるよ！",
		"なんでこんないっぱいになるまで放っておいたの！もう！",
	}},
}

// memPhrases はメモリ使用率に応じたセリフ
var memPhrases = []struct {
	threshold int
	phrases   []string
}{
	{50, []string{
		"メモリはまだ大丈夫やね。でも無駄遣いしたらあかんよ。",
		"今のうちに要らんアプリは閉めときなさいよ。",
	}},
	{75, []string{
		"メモリがだいぶ使われてるよ。大丈夫？要らんタブ閉めなさい！",
		"ブラウザのタブ何個開いてるの！もう！",
	}},
	{90, []string{
		"メモリがいっぱいになってきてるよ！タブ閉めなさい！",
		"メモリがパンパンやないの！何そんなに開いてるの！",
	}},
	{101, []string{
		"メモリがほぼいっぱいやないの！今すぐ何か閉めなさい！",
		"こんなにメモリ使って！パソコンが可哀想やで！",
	}},
}

// loadPhrases はロードアベレージに応じたセリフ（コア数比）
var loadPhrases = []struct {
	threshold float64
	phrases   []string
}{
	{0.5, []string{
		"CPUは余裕があるね。ゆっくり仕事しなさいよ。",
		"CPUも頑張ってるよ。あんたも頑張りなさいよ。",
	}},
	{1.0, []string{
		"CPUがまあまあ頑張ってるよ。無理させたらあかんよ。",
		"けっこうCPU使ってるね。大丈夫？",
	}},
	{2.0, []string{
		"CPUがだいぶしんどそうやで！要らんプロセスは止めなさい！",
		"CPUが悲鳴あげてるで！少し楽にしてあげなさいよ！",
	}},
	{1e9, []string{
		"CPUが火を噴きそうやないの！今すぐ何か止めなさい！",
		"こんなにCPU酷使して！パソコンが燃えるよ！",
	}},
}

// randomNags は脈絡なくいきなりガミガミしてくるセリフ
var randomNags = []string{
	"ちゃんと水飲んでる？こまめに飲まなあかんよ！",
	"背筋伸ばして！猫背になってるよ！",
	"目が疲れてへん？たまには遠くを見なさいよ。",
	"ちゃんと休憩取れてる？根詰めすぎたらあかんよ。",
	"手洗いした？ちゃんと手洗いしなさいよ！",
	"最近ちゃんと寝れてる？睡眠は大事やで！",
	"野菜食べてる？カップ麺ばっかりやないやろね？",
	"運動してる？たまには体動かしなさいよ！",
	"友達とちゃんと連絡取ってる？人間関係大事やで！",
	"お金の無駄遣いしてへん？貯金もしなさいよ！",
	"ストレス溜めてへん？悩みがあったらおかんに話しなさい！",
	"姿勢！姿勢！ちゃんと椅子に深く座りなさいよ！",
}

// cheers は応援セリフ
var cheers = []string{
	"頑張ってるやないの！おかん見てたよ、えらいえらい！",
	"あんたなら絶対できる！おかんが信じてるよ！",
	"しんどくても諦めたらあかんよ！頑張り！",
	"応援してるよ！何かあったらおかんに言いなさいよ！",
	"コツコツ頑張ってるやん！そういうの大事やで！",
}

// goodnights はおやすみセリフ
var goodnights = []string{
	"おやすみ！ちゃんと布団かぶって寝なさいよ。",
	"お疲れさん！今日も一日よう頑張ったね。おやすみ！",
	"早く寝なさいよ！明日もええ一日になるように祈ってるよ！",
	"おやすみなさい！ゆっくり休んでね。おかんも嬉しいよ。",
}

// goodmornings は「起きた」報告へのセリフ
var goodmornings = []string{
	"おはよう！ちゃんと起きれたやん！えらいえらい！",
	"おはようさん！今日もええ一日にしなさいよ！",
	"起きたんか！ほな朝ごはん食べて仕事しなさいよ！",
}

// Pick はスライスからランダムに1つ選ぶ
func Pick(phrases []string) string {
	if len(phrases) == 0 {
		return ""
	}
	return phrases[rand.Intn(len(phrases))]
}

// Greeting は現在の気分に合ったあいさつを返す
func Greeting() string {
	return Pick(greetings[CurrentMood()])
}

// RandomNag は脈絡のないガミガミを返す
func RandomNag() string {
	return Pick(randomNags)
}

// Cheer は応援メッセージを返す
func Cheer() string {
	return Pick(cheers)
}

// Goodnight はおやすみを返す
func Goodnight() string {
	return Pick(goodnights)
}

// GoodMorning はおはようを返す
func GoodMorning() string {
	return Pick(goodmornings)
}

// DiskComment はディスク使用率に応じたコメントを返す
func DiskComment(percent int) string {
	for _, p := range diskPhrases {
		if percent < p.threshold {
			return Pick(p.phrases)
		}
	}
	return Pick(diskPhrases[len(diskPhrases)-1].phrases)
}

// MemComment はメモリ使用率に応じたコメントを返す
func MemComment(percent int) string {
	for _, p := range memPhrases {
		if percent < p.threshold {
			return Pick(p.phrases)
		}
	}
	return Pick(memPhrases[len(memPhrases)-1].phrases)
}

// LoadComment はロードアベレージ比に応じたコメントを返す
func LoadComment(ratio float64) string {
	for _, p := range loadPhrases {
		if ratio < p.threshold {
			return Pick(p.phrases)
		}
	}
	return Pick(loadPhrases[len(loadPhrases)-1].phrases)
}
