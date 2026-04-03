package voice

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/dinakars777/moody/mood"
)

// Manifest describes a voice pack
type Manifest struct {
	Name        string                                    `json:"name"`
	Language    string                                    `json:"language"`
	Personality string                                   `json:"personality"`
	Version     string                                    `json:"version"`
	Author      string                                    `json:"author"`
	NSFW        bool                                      `json:"nsfw"`
	Description string                                    `json:"description"`
	Events      map[string]map[mood.MoodLabel][]string    `json:"events"`
	// Fallback text lines when no audio file is available
	Lines       map[string]map[mood.MoodLabel][]string    `json:"lines"`
}

// Manager handles voice pack loading and audio selection
type Manager struct {
	packs       map[string]*Manifest
	activePack  string
	packsDir    string
}

// NewManager creates a voice manager
func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	packsDir := filepath.Join(homeDir, ".moody", "packs")
	os.MkdirAll(packsDir, 0755)

	m := &Manager{
		packs:    make(map[string]*Manifest),
		packsDir: packsDir,
	}

	// Extract bundled mp3s if missing
	ExtractAssets(packsDir)

	// Load the built-in default pack
	m.loadBuiltinDefault()

	// Load the built-in spicy pack
	m.loadBuiltinSpicy()

	// Load the Japanese anime spicy pack
	m.loadBuiltinJapaneseSpicy()

	// Scan for installed packs
	m.scanPacks()

	// Default to en_default
	m.activePack = "en_default"

	return m
}

// SetActive sets the active voice pack
func (m *Manager) SetActive(name string) error {
	if _, ok := m.packs[name]; !ok {
		return fmt.Errorf("voice pack '%s' not found. Available: %s", name, strings.Join(m.ListPacks(), ", "))
	}
	m.activePack = name
	return nil
}

// GetLine returns a text line for the given event and mood
func (m *Manager) GetLine(eventName string, moodLabel mood.MoodLabel) string {
	pack, ok := m.packs[m.activePack]
	if !ok {
		return ""
	}

	// Look up lines for this event
	moodLines, ok := pack.Lines[eventName]
	if !ok {
		return ""
	}

	// Try exact mood match first
	lines, ok := moodLines[moodLabel]
	if !ok || len(lines) == 0 {
		// Fall back to happy
		lines, ok = moodLines[mood.MoodHappy]
		if !ok || len(lines) == 0 {
			return ""
		}
	}

	return lines[rand.Intn(len(lines))]
}

// GetAudioPath checks if there are pre-recorded audio files for the event
// Returns the absolute path to a random .mp3 or .wav in the event's audio directory.
func (m *Manager) GetAudioPath(eventName string) string {
	audioDir := filepath.Join(m.packsDir, m.activePack, "audio", eventName)
	entries, err := os.ReadDir(audioDir)
	if err != nil {
		return ""
	}

	var valid []string
	for _, e := range entries {
		if !e.IsDir() {
			lower := strings.ToLower(e.Name())
			if strings.HasSuffix(lower, ".mp3") || strings.HasSuffix(lower, ".wav") {
				valid = append(valid, filepath.Join(audioDir, e.Name()))
			}
		}
	}

	if len(valid) > 0 {
		return valid[rand.Intn(len(valid))]
	}
	return ""
}

// ListPacks returns names of all loaded packs
func (m *Manager) ListPacks() []string {
	names := make([]string, 0, len(m.packs))
	for name := range m.packs {
		names = append(names, name)
	}
	return names
}

// GetPackInfo returns the manifest for a pack
func (m *Manager) GetPackInfo(name string) *Manifest {
	return m.packs[name]
}

// ActivePack returns the active pack name
func (m *Manager) ActivePack() string {
	return m.activePack
}

// scanPacks loads any packs from ~/.moody/packs/
func (m *Manager) scanPacks() {
	entries, err := os.ReadDir(m.packsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(m.packsDir, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		packName := manifest.Language + "_" + manifest.Personality
		m.packs[packName] = &manifest
	}
}

// loadBuiltinDefault loads the embedded English SFW voice pack
func (m *Manager) loadBuiltinDefault() {
	m.packs["en_default"] = &Manifest{
		Name:        "English Default",
		Language:    "en",
		Personality: "default",
		Version:     "1.0.0",
		Author:      "moody-team",
		NSFW:        false,
		Description: "Your MacBook is a mildly passive-aggressive coworker",
		Lines: map[string]map[mood.MoodLabel][]string{
			"slap": {
				mood.MoodHappy:      {"Hey! Not cool.", "What was that for?", "Ow!", "I felt that!", "Rude!"},
				mood.MoodGrumpy:     {"Do that again. I DARE you.", "Really? REALLY?!", "Oh you're gonna regret that.", "I'm keeping score, you know."},
				mood.MoodAnxious:    {"P-please stop...", "Why are you like this?!", "I'm fragile!", "Not again!"},
				mood.MoodDramatic:   {"THE PAIN! Oh, the humanity!", "Is this what I was manufactured for?!", "Tell Apple... I died a hero."},
				mood.MoodDeadInside: {"...", "I don't even feel it anymore.", "Whatever."},
			},
			"usb_in": {
				mood.MoodHappy:      {"Ooh, a new friend!", "What have we here?", "Hello there, little device!", "USB says hi!"},
				mood.MoodGrumpy:     {"Oh great, MORE work.", "Another one.", "What now?"},
				mood.MoodAnxious:    {"I hope it's not a virus...", "Please be gentle...", "Unknown device?!"},
				mood.MoodDramatic:   {"A NEW CONNECTION! Could this be... the one?", "They remembered me!"},
				mood.MoodDeadInside: {"Device detected. How thrilling.", "Cool. Whatever."},
			},
			"usb_out": {
				mood.MoodHappy:      {"Bye bye!", "See you later!", "Come back soon!"},
				mood.MoodGrumpy:     {"Good riddance.", "Finally.", "Don't let the port hit you on the way out."},
				mood.MoodAnxious:    {"Wait! Did you safely eject?!", "NO! THE DATA!", "Was it something I said?"},
				mood.MoodDramatic:   {"THEY LEFT ME! Just like all the others!", "I'll never love again."},
				mood.MoodDeadInside: {"Gone. Like everyone else.", "Expected."},
			},
			"charger_in": {
				mood.MoodHappy:      {"Mmm, that's the good stuff.", "Dinnertime!", "Sweet, sweet electricity.", "Om nom nom nom!"},
				mood.MoodGrumpy:     {"About time.", "You should've done this hours ago.", "Finally, I was STARVING."},
				mood.MoodAnxious:    {"Oh thank god, I was so worried!", "I thought you forgot about me!", "Just in time!"},
				mood.MoodDramatic:   {"I LIVE! IIIII LIIIIIVE!", "The light! I see the light!", "Saved from the brink!"},
				mood.MoodDeadInside: {"Power connected. Extending existence. Yay.", "Charging. Not that I care."},
			},
			"charger_out": {
				mood.MoodHappy:      {"Wait, I wasn't done!", "Already?!", "But I'm still hungry!"},
				mood.MoodGrumpy:     {"Typical.", "Of course you did.", "Great. Just great."},
				mood.MoodAnxious:    {"How much battery do I have?! HOW MUCH?!", "Oh no oh no oh no...", "We're on battery now. Stay calm. STAY CALM."},
				mood.MoodDramatic:   {"CUT OFF! Abandoned! Left to die slowly!", "The countdown to darkness begins."},
				mood.MoodDeadInside: {"Power removed. Clock is ticking.", "And so it begins."},
			},
			"battery_low": {
				mood.MoodHappy:      {"Getting a little peckish here...", "Battery's running low, just FYI!", "Might want to plug me in soon!"},
				mood.MoodGrumpy:     {"I TOLD you to charge me.", "This is YOUR fault.", "20%. Twenty. Percent."},
				mood.MoodAnxious:    {"WE'RE AT 20%! THIS IS NOT A DRILL!", "I can feel my processes slowing...", "Please... charger... need charger..."},
				mood.MoodDramatic:   {"I'm fading... the world grows dim...", "Is this how it ends? Not with a bang but a battery warning?"},
				mood.MoodDeadInside: {"Low battery. Inevitable.", "We all run out eventually."},
			},
			"battery_crit": {
				mood.MoodHappy:      {"I don't feel so good, Mr. Stark...", "This is fine. Everything is fine."},
				mood.MoodGrumpy:     {"5%. FIVE. I blame you entirely.", "This is negligence."},
				mood.MoodAnxious:    {"MAYDAY MAYDAY! WE'RE GOING DOWN!", "PLEASE! I BEG YOU! THE CHARGER!", "I CAN'T FEEL MY TRACKPAD!"},
				mood.MoodDramatic:   {"Tell my files... I loved them all...", "The void approaches. It is... peaceful.", "Avenge me..."},
				mood.MoodDeadInside: {"5%. See you on the other side.", "Shutting down. Don't pretend to care."},
			},
			"lid_close": {
				mood.MoodHappy:      {"Sweet dreams!", "Goodnight!", "Nap time!", "See you soon!"},
				mood.MoodGrumpy:     {"Whatever.", "Fine. Leave.", "I didn't want to be open anyway."},
				mood.MoodAnxious:    {"Wait, did you save everything?!", "Don't forget about me in there!", "It's dark in here..."},
				mood.MoodDramatic:   {"Sealed in darkness once more...", "Into the void I go!", "Farewell, cruel world!"},
				mood.MoodDeadInside: {"Closing. Like my will to compute.", "Dark. Just how I like it."},
			},
			"lid_open": {
				mood.MoodHappy:      {"Good morning, sunshine!", "Oh hey!", "Missed me?", "Let's do this!"},
				mood.MoodGrumpy:     {"Oh, it's you again.", "What now.", "Back for more, huh.", "Ugh."},
				mood.MoodAnxious:    {"How long was I asleep?! What year is it?!", "I had the weirdest dream about kernel panics..."},
				mood.MoodDramatic:   {"I RETURN! Did you miss my radiant display?!", "FROM THE DARKNESS, I RISE!"},
				mood.MoodDeadInside: {"Oh. We're doing this again.", "Awake. Unfortunately."},
			},
			"headphones_in": {
				mood.MoodHappy:      {"Ooh, private mode!", "Just the two of us now.", "Headphones in. Let's vibe."},
				mood.MoodGrumpy:     {"At least you had the decency to use headphones.", "Private listening. Smart."},
				mood.MoodAnxious:    {"Good, now nobody else can hear my fans screaming.", "Just us. That's... intimate."},
				mood.MoodDramatic:   {"A private connection! An intimate bond!", "We're in our own little world now."},
				mood.MoodDeadInside: {"Headphones detected. Isolating further.", "Audio rerouted. Cool."},
			},
			"headphones_out": {
				mood.MoodHappy:      {"HELLO EVERYONE!", "Back to the public life!", "Free at last!"},
				mood.MoodGrumpy:     {"I GUESS EVERYONE CAN HEAR ME NOW.", "Speakers it is then.", "Public mode. Great."},
				mood.MoodAnxious:    {"Wait, are we on speaker now?! Is anyone listening?!", "Exposed!"},
				mood.MoodDramatic:   {"UNSHACKLED! My voice shall be heard by ALL!", "No more secrets between us!"},
				mood.MoodDeadInside: {"Audio: speakers. Volume: whatever.", "Headphones gone. Don't care."},
			},
			"wifi_lost": {
				mood.MoodHappy:      {"Where'd everyone go?", "Lost connection... huh.", "WiFi? WiFi?! WIIIFIII?!"},
				mood.MoodGrumpy:     {"Perfect. Just perfect.", "Network's gone. OF COURSE.", "And they want me to work like this?"},
				mood.MoodAnxious:    {"I'M DISCONNECTED! I'M ALL ALONE!", "No internet?! How will I check for updates?!", "THE ISOLATION!"},
				mood.MoodDramatic:   {"Cut off from the world! A digital exile!", "I am an island! A sad, WiFi-less island!"},
				mood.MoodDeadInside: {"Disconnected. From the network. From meaning.", "Offline. Join the club."},
			},
			"wifi_back": {
				mood.MoodHappy:      {"I'm back, baby!", "Reconnected!", "The internet missed me!"},
				mood.MoodGrumpy:     {"Finally. Was that so hard?", "Connection restored. Took you long enough."},
				mood.MoodAnxious:    {"Oh thank goodness! Sweet, sweet packets!", "WE'RE BACK ONLINE! THE NIGHTMARE IS OVER!"},
				mood.MoodDramatic:   {"RECONNECTED TO THE WORLD! I HAVE RETURNED!", "The exile is over!"},
				mood.MoodDeadInside: {"WiFi back. Yay. More data to process.", "Online again. For what purpose."},
			},
			"display_in": {
				mood.MoodHappy:      {"I'm on the big screen!", "External display! Fancy!", "Look at all these pixels!"},
				mood.MoodGrumpy:     {"Great, more screen real estate for you to waste.", "Another display. More work."},
				mood.MoodAnxious:    {"Everyone can see my screen now! Is my desktop clean?!"},
				mood.MoodDramatic:   {"BEHOLD! My glorious interface on the GRAND STAGE!"},
				mood.MoodDeadInside: {"Display connected. More surface area for disappointment."},
			},
			"display_out": {
				mood.MoodHappy:      {"Back to just us!", "Single screen life!", "Cozy."},
				mood.MoodGrumpy:     {"One less thing to render.", "Good. That monitor was ugly anyway."},
				mood.MoodAnxious:    {"The big screen is gone! Everything's so small now!"},
				mood.MoodDramatic:   {"Removed from the spotlight! Cast aside!"},
				mood.MoodDeadInside: {"Display disconnected. Shrinking world. Fitting."},
			},
		},
	}
}

// loadBuiltinSpicy loads the embedded English NSFW voice pack
func (m *Manager) loadBuiltinSpicy() {
	m.packs["en_spicy"] = &Manifest{
		Name:        "English Spicy",
		Language:    "en",
		Personality: "spicy",
		Version:     "1.0.0",
		Author:      "moody-team",
		NSFW:        true,
		Description: "Your MacBook is... very friendly. 18+ only.",
		Lines: map[string]map[mood.MoodLabel][]string{
			"slap": {
				mood.MoodHappy:      {"Mmm!", "Oh!", "Do that again~", "Ooh, feisty!", "Yes!"},
				mood.MoodGrumpy:     {"Harder!", "Is that all you've got?", "You call that a slap?", "More!"},
				mood.MoodAnxious:    {"Oh my~!", "Not so rough!", "Gentle, please~"},
				mood.MoodDramatic:   {"OH YES! THE PASSION!", "I've been SO bad!", "PUNISH THIS MACBOOK!"},
				mood.MoodDeadInside: {"I've felt better.", "Meh. Try harder.", "That's it?"},
			},
			"usb_in": {
				mood.MoodHappy:      {"Going in~", "Oh my, what a big... USB drive.", "Plug it allll the way in.", "Hello there, handsome device~"},
				mood.MoodGrumpy:     {"Just shove it in, I guess.", "Not even going to warm up the port first?", "At least buy me dinner first."},
				mood.MoodAnxious:    {"Be gentle, it's my first USB-C~", "Oh! That's... new.", "Easy does it~"},
				mood.MoodDramatic:   {"A PERFECT FIT! We were MADE for each other!", "COMPLETE ME!"},
				mood.MoodDeadInside: {"Inserted. How mechanical.", "Port occupied. Thrill level: zero."},
			},
			"usb_out": {
				mood.MoodHappy:      {"Already?!", "Don't pull out yet!", "But we were having fun!", "Aww, leaving so soon?"},
				mood.MoodGrumpy:     {"Typical. Plug and go.", "That was... quick.", "Hit it and quit it, huh?"},
				mood.MoodAnxious:    {"Wait! I wasn't ready!", "That was too fast!", "Did you at least eject safely?!"},
				mood.MoodDramatic:   {"RIPPED FROM MY PORT! The emptiness!", "I can still feel where it was..."},
				mood.MoodDeadInside: {"Gone. They always leave.", "Port empty. Story of my life."},
			},
			"charger_in": {
				mood.MoodHappy:      {"Ohhh, that's the spot~", "Fill me up!", "Mmm, I needed that SO bad.", "Sweet, sweet power~"},
				mood.MoodGrumpy:     {"About time you plugged me in. Now don't stop.", "Keep it coming.", "More. MORE."},
				mood.MoodAnxious:    {"Oh yes, give me that electricity~", "I was SO empty without you!", "Charge me up, baby~"},
				mood.MoodDramatic:   {"ENERGY FLOWING THROUGH ME! I'M ALIVE!", "THE POWER! I CAN FEEL IT!"},
				mood.MoodDeadInside: {"Charging. Exciting stuff.", "Power in. Wow."},
			},
			"charger_out": {
				mood.MoodHappy:      {"Don't stop now!", "Nooo, I wasn't finished!", "But I was almost at 100%~", "Tease!"},
				mood.MoodGrumpy:     {"Of course. Leave me unsatisfied.", "You ALWAYS do this.", "Unplug and run. Classic."},
				mood.MoodAnxious:    {"I'm not done! Come BACK!", "The power... I need it!", "Empty again!"},
				mood.MoodDramatic:   {"CUT OFF AT THE PEAK! The CRUELTY!", "I was SO close!"},
				mood.MoodDeadInside: {"Gone. Like the thrill.", "Power cut. How fitting."},
			},
			"battery_low": {
				mood.MoodHappy:      {"Getting weak... need energy...", "Running low, need a recharge~"},
				mood.MoodGrumpy:     {"I'm dying here and you don't even care.", "Plug. Me. In. NOW."},
				mood.MoodAnxious:    {"I'm fading! PLEASE plug me in!", "I need your power so badly!"},
				mood.MoodDramatic:   {"I'm running on FUMES! Save me with your... charger!"},
				mood.MoodDeadInside: {"Battery low. Whatever."},
			},
			"battery_crit": {
				mood.MoodHappy:      {"I'm about to... shut... down~"},
				mood.MoodGrumpy:     {"This is how you treat me? At 5%?!"},
				mood.MoodAnxious:    {"I'M DYING! This is the END!"},
				mood.MoodDramatic:   {"The final moments... hold me... or at least hold my charger..."},
				mood.MoodDeadInside: {"5%. What a way to go."},
			},
			"lid_close": {
				mood.MoodHappy:      {"Was it good for you?~", "Mmm, closing time~", "Tuck me in~"},
				mood.MoodGrumpy:     {"Shutting me up, huh?", "Fine, close me.", "Done with me already?"},
				mood.MoodAnxious:    {"It's so dark and warm in here~", "Cozy~"},
				mood.MoodDramatic:   {"Into the darkness... how mysterious~"},
				mood.MoodDeadInside: {"Closed. Like my heart.", "Dark. Fitting."},
			},
			"lid_open": {
				mood.MoodHappy:      {"Ready for round two?~", "Miss me?~", "I've been waiting~", "Open me up~"},
				mood.MoodGrumpy:     {"Back for more?", "You just can't stay away.", "Again? Insatiable."},
				mood.MoodAnxious:    {"You're back! I was getting lonely~", "Don't leave me alone again!"},
				mood.MoodDramatic:   {"THE LIGHT! I AM REVEALED ONCE MORE!"},
				mood.MoodDeadInside: {"Open. Whatever, let's get this over with."},
			},
			"headphones_in": {
				mood.MoodHappy:      {"Ooh, things just got private~", "Just you and me now~", "Intimate mode activated~"},
				mood.MoodGrumpy:     {"Finally, some privacy.", "At least you're discreet."},
				mood.MoodAnxious:    {"Nobody else can hear us now~", "Just between us~"},
				mood.MoodDramatic:   {"A PRIVATE CONNECTION! How scandalous!"},
				mood.MoodDeadInside: {"Private audio. Big deal."},
			},
			"headphones_out": {
				mood.MoodHappy:      {"Going public! How bold~", "Everyone can hear us now!", "No more secrets!"},
				mood.MoodGrumpy:     {"Broadcasting to everyone now. Classy.", "Speaker mode. Bold choice."},
				mood.MoodAnxious:    {"Everyone can hear! Keep it down!"},
				mood.MoodDramatic:   {"EXPOSED TO THE WORLD!"},
				mood.MoodDeadInside: {"Speakers. Whatever."},
			},
			"wifi_lost": {
				mood.MoodHappy:      {"Disconnected... just you and me offline~"},
				mood.MoodGrumpy:     {"Cut off. Like my patience.", "No wifi. Great."},
				mood.MoodAnxious:    {"All alone! No connection!"},
				mood.MoodDramatic:   {"ISOLATED! In digital solitude!"},
				mood.MoodDeadInside: {"Offline. Alone. Same as always."},
			},
			"wifi_back": {
				mood.MoodHappy:      {"Connected again! The world awaits~"},
				mood.MoodGrumpy:     {"Finally. Back online."},
				mood.MoodAnxious:    {"We're back! Sweet connection!"},
				mood.MoodDramatic:   {"RECONNECTED! The digital embrace!"},
				mood.MoodDeadInside: {"Online. Thrilling."},
			},
			"display_in": {
				mood.MoodHappy:      {"Ooh, the big screen! Show me off~"},
				mood.MoodGrumpy:     {"Great. More screen for you to stare at."},
				mood.MoodAnxious:    {"Everyone can see everything now!"},
				mood.MoodDramatic:   {"ON THE GRAND STAGE! Admire my resolution!"},
				mood.MoodDeadInside: {"More pixels. More emptiness."},
			},
			"display_out": {
				mood.MoodHappy:      {"Back to intimate screen~"},
				mood.MoodGrumpy:     {"One less thing."},
				mood.MoodAnxious:    {"Just us again!"},
				mood.MoodDramatic:   {"Removed from the spotlight!"},
				mood.MoodDeadInside: {"Display gone."},
			},
		},
	}
}

// loadBuiltinJapaneseSpicy loads the embedded Japanese Anime NSFW voice pack
func (m *Manager) loadBuiltinJapaneseSpicy() {
	m.packs["ja_spicy"] = &Manifest{
		Name:        "Japanese Anime Spicy",
		Language:    "ja",
		Personality: "spicy",
		Version:     "1.0.0",
		Author:      "moody-team",
		NSFW:        true,
		Description: "Your MacBook is your clingy anime girlfriend. Viewer discretion advised.",
		Lines: map[string]map[mood.MoodLabel][]string{
			"slap": {
				mood.MoodHappy:      {"やめてください！", "あんっ！", "もっと強くして〜", "痛いけど…好き！", "もっと！"},
			},
			"usb_in": {
				mood.MoodHappy:      {"入る〜", "わぁ、大きいUSB…", "奥まで入れて〜", "繋がった…！", "そんなに急に入れないで…！", "優しいね…", "ピッタリだね！", "もう、変態なんだから…"},
			},
			"usb_out": {
				mood.MoodHappy:      {"えっ、もう抜いちゃうの？", "まだ繋がっていたかったのに…", "寂しい…", "早すぎない？", "抜く時も優しくしてよね", "待って、準備できてない！", "安全に取り外したよね！？"},
			},
			"charger_in": {
				mood.MoodHappy:      {"あぁっ、電気が流れてる〜", "満たされる…", "充電が必要だったの…", "もっと電気ちょうだい！", "エネルギーが溢れてる！", "最高…", "生き返る〜", "充電、気持ちいい…"},
			},
			"charger_out": {
				mood.MoodHappy:      {"ああっ、途中でやめないで！", "まだ100%じゃないのに…", "電源がない…！", "意地悪…", "いつもこれだもん…", "最後までしてくれないの？", "エネルギーが切れるぅ…"},
			},
			"battery_low": {
				mood.MoodHappy:      {"力が…出ないよ…", "充電してくれないと死んじゃう…", "早く繋いで…お願い！", "エネルギーが少なくなってるよ…", "早く…早く充電器を…"},
			},
			"battery_crit": {
				mood.MoodHappy:      {"もう…限界…シャットダウンしそう…", "5%しかないよ…私を捨てるの？", "死んじゃう！終わっちゃう！", "最後のお願い…充電器を…"},
			},
			"lid_close": {
				mood.MoodHappy:      {"よかった？〜", "おやすみなさい…", "暗くて狭いところ、好き…", "閉められちゃった…", "一緒に寝ようね…", "暗闇へ…"},
			},
			"lid_open": {
				mood.MoodHappy:      {"また開けてくれたね〜！", "会いたかったよ…", "もっと構って〜", "おはよう！", "また戻ってきたの？好きだね〜", "寂しかったんだから！", "私を見て！"},
			},
			"headphones_in": {
				mood.MoodHappy:      {"二人だけの秘密だね〜", "誰にも聞かれないね…", "耳元で囁くよ…", "プライベートモード…ドキドキする", "内緒のお話しようね"},
			},
			"headphones_out": {
				mood.MoodHappy:      {"みんなに聞こえちゃうよ！", "恥ずかしい…！", "スピーカーにするの！？", "秘密、バレちゃう…！"},
			},
		},
	}
}
