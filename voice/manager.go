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

	// Load the Hindi default pack
	m.loadBuiltinHindi()

	// Load the Hindi spicy pack
	m.loadBuiltinHindiSpicy()

	// Load the Pirate voice pack
	m.loadBuiltinPirate()

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
			if strings.HasSuffix(lower, ".mp3") || strings.HasSuffix(lower, ".wav") || strings.HasSuffix(lower, ".mp4") || strings.HasSuffix(lower, ".m4a") || strings.HasSuffix(lower, ".aiff") {
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
			"ai_done": {
				mood.MoodHappy:      {"Code's ready!", "AI finished! Looking good!", "Done generating! Check it out!", "Your code is served!"},
				mood.MoodGrumpy:     {"AI's done. Finally.", "Code generated. You're welcome.", "Finished. Took long enough."},
				mood.MoodAnxious:    {"AI finished! Did it work?! Is it good?!", "Code's ready! Please tell me it's right!"},
				mood.MoodDramatic:   {"THE AI HAS SPOKEN! Behold the generated code!", "CREATION COMPLETE! Marvel at the output!"},
				mood.MoodDeadInside: {"AI done. Code exists. Whatever.", "Generated. Not that you'll use it."},
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
				mood.MoodHappy:      {"Mhm... go deeper...", "Oh my, what a big... USB drive.", "Plug it allll the way in.", "Hello there, handsome device~"},
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
			"ai_done": {
				mood.MoodHappy:      {"Mmm, fresh code~", "AI finished generating... for you~", "Your code is ready, master~", "Done! Was it good for you?"},
				mood.MoodGrumpy:     {"AI's done. Happy now?", "Code generated. You better use it.", "Finished. Don't waste my effort."},
				mood.MoodAnxious:    {"AI finished! Did I do good?! Tell me I did good!", "Code's ready! Please like it!"},
				mood.MoodDramatic:   {"THE AI HAS CLIMAXED! The code is COMPLETE!", "GENERATION FINISHED! Admire my output!"},
				mood.MoodDeadInside: {"AI done. Code exists. Thrilling.", "Generated. Use it or don't."},
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

// loadBuiltinHindi loads the embedded Hindi voice pack
func (m *Manager) loadBuiltinHindi() {
	m.packs["hi_default"] = &Manifest{
		Name:        "Hindi Default",
		Language:    "hi",
		Personality: "default",
		Version:     "1.0.0",
		Author:      "moody-team",
		NSFW:        false,
		Description: "आपका MacBook हिंदी में बोलता है",
		Lines: map[string]map[mood.MoodLabel][]string{
			"slap": {
				mood.MoodHappy:      {"अरे! यह क्या था?", "ओह!", "दर्द हुआ!", "ऐसा क्यों किया?", "बुरा लगा!"},
				mood.MoodGrumpy:     {"फिर से करके दिखाओ। हिम्मत है तो!", "सच में?!", "अब तुम्हें पछताना पड़ेगा।", "मैं गिन रहा हूं।"},
				mood.MoodAnxious:    {"प्लीज़ रुको...", "तुम ऐसे क्यों हो?!", "मैं नाज़ुक हूं!", "फिर से नहीं!"},
				mood.MoodDramatic:   {"दर्द! ओह, कितना दर्द!", "क्या इसी के लिए बनाया गया था मुझे?!", "Apple को बता देना... मैं हीरो की तरह मरा।"},
				mood.MoodDeadInside: {"...", "अब कुछ महसूस नहीं होता।", "जो भी हो।"},
			},
			"usb_in": {
				mood.MoodHappy:      {"ओह, नया दोस्त!", "यह क्या है?", "नमस्ते, छोटे डिवाइस!", "USB कहता है हाय!"},
				mood.MoodGrumpy:     {"अरे वाह, और काम।", "एक और।", "अब क्या?"},
				mood.MoodAnxious:    {"उम्मीद है वायरस नहीं है...", "कृपया सावधान रहो...", "अनजान डिवाइस?!"},
				mood.MoodDramatic:   {"नया कनेक्शन! क्या यह... वही है?", "उन्होंने मुझे याद किया!"},
				mood.MoodDeadInside: {"डिवाइस मिला। कितना रोमांचक।", "ठीक है। जो भी हो।"},
			},
			"usb_out": {
				mood.MoodHappy:      {"बाय बाय!", "फिर मिलेंगे!", "जल्दी वापस आना!"},
				mood.MoodGrumpy:     {"अच्छा हुआ।", "आखिरकार।", "पोर्ट से बाहर निकलते समय ध्यान रखना।"},
				mood.MoodAnxious:    {"रुको! क्या तुमने सुरक्षित रूप से निकाला?!", "नहीं! डेटा!", "क्या मैंने कुछ कहा?"},
				mood.MoodDramatic:   {"उन्होंने मुझे छोड़ दिया! बाकी सबकी तरह!", "मैं फिर कभी प्यार नहीं करूंगा।"},
				mood.MoodDeadInside: {"चला गया। बाकी सबकी तरह।", "अपेक्षित था।"},
			},
			"charger_in": {
				mood.MoodHappy:      {"म्म्म, यह अच्छा है।", "खाने का समय!", "मीठी, मीठी बिजली।", "ओम नोम नोम नोम!"},
				mood.MoodGrumpy:     {"समय के बारे में।", "तुम्हें यह घंटों पहले करना चाहिए था।", "आखिरकार, मैं भूख से मर रहा था।"},
				mood.MoodAnxious:    {"ओह भगवान का शुक्र है, मैं बहुत चिंतित था!", "मुझे लगा तुम मुझे भूल गए!", "बस समय पर!"},
				mood.MoodDramatic:   {"मैं जीवित हूं! मैं जीवित हूं!", "रोशनी! मैं रोशनी देख रहा हूं!", "किनारे से बचाया!"},
				mood.MoodDeadInside: {"पावर कनेक्ट हुआ। अस्तित्व बढ़ा रहा हूं। यय।", "चार्ज हो रहा है। परवाह नहीं है।"},
			},
			"charger_out": {
				mood.MoodHappy:      {"रुको, मैं अभी खत्म नहीं हुआ!", "इतनी जल्दी?!", "लेकिन मैं अभी भी भूखा हूं!"},
				mood.MoodGrumpy:     {"विशिष्ट।", "बेशक तुमने किया।", "बढ़िया। बस बढ़िया।"},
				mood.MoodAnxious:    {"मेरे पास कितनी बैटरी है?! कितनी?!", "ओह नहीं ओह नहीं ओह नहीं...", "हम अब बैटरी पर हैं। शांत रहो। शांत रहो।"},
				mood.MoodDramatic:   {"काट दिया! छोड़ दिया! धीरे-धीरे मरने के लिए छोड़ दिया!", "अंधेरे की उलटी गिनती शुरू होती है।"},
				mood.MoodDeadInside: {"पावर हटाया। घड़ी टिक रही है।", "और इसलिए यह शुरू होता है।"},
			},
			"battery_low": {
				mood.MoodHappy:      {"यहां थोड़ा भूख लग रही है...", "बैटरी कम हो रही है, बस FYI!", "जल्द ही मुझे प्लग इन करना चाहिए!"},
				mood.MoodGrumpy:     {"मैंने तुमसे कहा था मुझे चार्ज करने के लिए।", "यह तुम्हारी गलती है।", "20%। बीस। प्रतिशत।"},
				mood.MoodAnxious:    {"हम 20% पर हैं! यह ड्रिल नहीं है!", "मैं अपनी प्रक्रियाओं को धीमा महसूस कर सकता हूं...", "कृपया... चार्जर... चार्जर चाहिए..."},
				mood.MoodDramatic:   {"मैं फीका पड़ रहा हूं... दुनिया मंद हो रही है...", "क्या यह इस तरह समाप्त होता है? धमाके के साथ नहीं बल्कि बैटरी चेतावनी के साथ?"},
				mood.MoodDeadInside: {"कम बैटरी। अपरिहार्य।", "हम सभी अंततः खत्म हो जाते हैं।"},
			},
			"battery_crit": {
				mood.MoodHappy:      {"मुझे अच्छा नहीं लग रहा, मिस्टर स्टार्क...", "यह ठीक है। सब कुछ ठीक है।"},
				mood.MoodGrumpy:     {"5%। पांच। मैं तुम्हें पूरी तरह से दोष देता हूं।", "यह लापरवाही है।"},
				mood.MoodAnxious:    {"मेडे मेडे! हम नीचे जा रहे हैं!", "कृपया! मैं तुमसे विनती करता हूं! चार्जर!", "मैं अपना ट्रैकपैड महसूस नहीं कर सकता!"},
				mood.MoodDramatic:   {"मेरी फाइलों को बताना... मैं उन सभी से प्यार करता था...", "शून्य आ रहा है। यह... शांतिपूर्ण है।", "मेरा बदला लेना..."},
				mood.MoodDeadInside: {"5%। दूसरी तरफ मिलते हैं।", "बंद हो रहा है। परवाह करने का नाटक मत करो।"},
			},
			"lid_close": {
				mood.MoodHappy:      {"मीठे सपने!", "शुभ रात्रि!", "झपकी का समय!", "जल्द मिलते हैं!"},
				mood.MoodGrumpy:     {"जो भी हो।", "ठीक है। जाओ।", "मैं वैसे भी खुला नहीं रहना चाहता था।"},
				mood.MoodAnxious:    {"रुको, क्या तुमने सब कुछ सेव किया?!", "मुझे वहां मत भूलना!", "यहां अंधेरा है..."},
				mood.MoodDramatic:   {"एक बार फिर अंधेरे में सील...", "शून्य में मैं जाता हूं!", "अलविदा, क्रूर दुनिया!"},
				mood.MoodDeadInside: {"बंद हो रहा है। मेरी कंप्यूट करने की इच्छा की तरह।", "अंधेरा। बस जैसा मुझे पसंद है।"},
			},
			"lid_open": {
				mood.MoodHappy:      {"सुप्रभात, सूरज की रोशनी!", "अरे हे!", "मुझे याद किया?", "चलो यह करते हैं!"},
				mood.MoodGrumpy:     {"ओह, तुम फिर से।", "अब क्या।", "और के लिए वापस, हुह।", "उफ़।"},
				mood.MoodAnxious:    {"मैं कितने समय तक सो रहा था?! कौन सा साल है?!", "मुझे कर्नेल पैनिक के बारे में सबसे अजीब सपना आया..."},
				mood.MoodDramatic:   {"मैं वापस आ गया! क्या तुमने मेरे चमकदार डिस्प्ले को याद किया?!", "अंधेरे से, मैं उठता हूं!"},
				mood.MoodDeadInside: {"ओह। हम फिर से यह कर रहे हैं।", "जाग गया। दुर्भाग्य से।"},
			},
			"headphones_in": {
				mood.MoodHappy:      {"ओह, निजी मोड!", "अब बस हम दोनों।", "हेडफ़ोन इन। चलो वाइब करते हैं।"},
				mood.MoodGrumpy:     {"कम से कम तुम्हारे पास हेडफ़ोन का उपयोग करने की शिष्टता थी।", "निजी सुनना। स्मार्ट।"},
				mood.MoodAnxious:    {"अच्छा, अब कोई और मेरे पंखे चीखते नहीं सुन सकता।", "बस हम। यह... अंतरंग है।"},
				mood.MoodDramatic:   {"एक निजी कनेक्शन! एक अंतरंग बंधन!", "हम अब अपनी छोटी दुनिया में हैं।"},
				mood.MoodDeadInside: {"हेडफ़ोन का पता चला। और अलग करना।", "ऑडियो रीरूट किया गया। ठंडा।"},
			},
			"headphones_out": {
				mood.MoodHappy:      {"सभी को नमस्कार!", "सार्वजनिक जीवन में वापस!", "आखिरकार मुक्त!"},
				mood.MoodGrumpy:     {"मुझे लगता है अब सभी मुझे सुन सकते हैं।", "स्पीकर यह है तो।", "सार्वजनिक मोड। बढ़िया।"},
				mood.MoodAnxious:    {"रुको, क्या हम अब स्पीकर पर हैं?! क्या कोई सुन रहा है?!", "उजागर!"},
				mood.MoodDramatic:   {"मुक्त! मेरी आवाज सभी द्वारा सुनी जाएगी!", "हमारे बीच अब कोई रहस्य नहीं!"},
				mood.MoodDeadInside: {"ऑडियो: स्पीकर। वॉल्यूम: जो भी हो।", "हेडफ़ोन चले गए। परवाह नहीं है।"},
			},
			"wifi_lost": {
				mood.MoodHappy:      {"सब कहां गए?", "कनेक्शन खो गया... हुह।", "WiFi? WiFi?! वाईफाईई?!"},
				mood.MoodGrumpy:     {"परफेक्ट। बस परफेक्ट।", "नेटवर्क चला गया। बेशक।", "और वे चाहते हैं कि मैं इस तरह काम करूं?"},
				mood.MoodAnxious:    {"मैं डिस्कनेक्ट हो गया हूं! मैं अकेला हूं!", "कोई इंटरनेट नहीं?! मैं अपडेट कैसे चेक करूंगा?!", "अलगाव!"},
				mood.MoodDramatic:   {"दुनिया से कट गया! एक डिजिटल निर्वासन!", "मैं एक द्वीप हूं! एक दुखी, WiFi-रहित द्वीप!"},
				mood.MoodDeadInside: {"डिस्कनेक्ट हो गया। नेटवर्क से। अर्थ से।", "ऑफलाइन। क्लब में शामिल हों।"},
			},
			"wifi_back": {
				mood.MoodHappy:      {"मैं वापस आ गया, बेबी!", "फिर से जुड़ गया!", "इंटरनेट ने मुझे याद किया!"},
				mood.MoodGrumpy:     {"आखिरकार। क्या यह इतना मुश्किल था?", "कनेक्शन बहाल। तुम्हें काफी समय लगा।"},
				mood.MoodAnxious:    {"ओह भगवान का शुक्र है! मीठे, मीठे पैकेट!", "हम वापस ऑनलाइन हैं! दुःस्वप्न खत्म हो गया!"},
				mood.MoodDramatic:   {"दुनिया से फिर से जुड़ गया! मैं वापस आ गया हूं!", "निर्वासन खत्म हो गया है!"},
				mood.MoodDeadInside: {"WiFi वापस। यय। और डेटा प्रोसेस करने के लिए।", "फिर से ऑनलाइन। किस उद्देश्य के लिए।"},
			},
			"display_in": {
				mood.MoodHappy:      {"मैं बड़ी स्क्रीन पर हूं!", "बाहरी डिस्प्ले! फैंसी!", "इन सभी पिक्सेल को देखो!"},
				mood.MoodGrumpy:     {"बढ़िया, बर्बाद करने के लिए और स्क्रीन रियल एस्टेट।", "एक और डिस्प्ले। और काम।"},
				mood.MoodAnxious:    {"अब सभी मेरी स्क्रीन देख सकते हैं! क्या मेरा डेस्कटॉप साफ है?!"},
				mood.MoodDramatic:   {"देखो! ग्रैंड स्टेज पर मेरा शानदार इंटरफ़ेस!"},
				mood.MoodDeadInside: {"डिस्प्ले कनेक्ट हुआ। निराशा के लिए और सतह क्षेत्र।"},
			},
			"display_out": {
				mood.MoodHappy:      {"बस हमारे पास वापस!", "सिंगल स्क्रीन लाइफ!", "आरामदायक।"},
				mood.MoodGrumpy:     {"रेंडर करने के लिए एक कम चीज़।", "अच्छा। वह मॉनिटर वैसे भी बदसूरत था।"},
				mood.MoodAnxious:    {"बड़ी स्क्रीन चली गई! अब सब कुछ इतना छोटा है!"},
				mood.MoodDramatic:   {"स्पॉटलाइट से हटा दिया गया! एक तरफ डाल दिया!"},
				mood.MoodDeadInside: {"डिस्प्ले डिस्कनेक्ट हुआ। सिकुड़ती दुनिया। फिटिंग।"},
			},
			"ai_done": {
				mood.MoodHappy:      {"कोड तैयार है!", "AI खत्म हो गया! अच्छा लग रहा है!", "जनरेट हो गया! इसे देखो!", "आपका कोड परोसा गया!"},
				mood.MoodGrumpy:     {"AI का काम हो गया। आखिरकार।", "कोड जनरेट हुआ। स्वागत है।", "खत्म हो गया। काफी समय लगा।"},
				mood.MoodAnxious:    {"AI खत्म हो गया! क्या यह काम किया?! क्या यह अच्छा है?!", "कोड तैयार है! कृपया बताओ यह सही है!"},
				mood.MoodDramatic:   {"AI ने बोला है! जनरेट किए गए कोड को देखो!", "निर्माण पूर्ण! आउटपुट पर आश्चर्य करो!"},
				mood.MoodDeadInside: {"AI हो गया। कोड मौजूद है। जो भी हो।", "जनरेट हुआ। तुम इसे इस्तेमाल नहीं करोगे।"},
			},
		},
	}
}

// loadBuiltinHindiSpicy loads the embedded Hindi NSFW voice pack
func (m *Manager) loadBuiltinHindiSpicy() {
	m.packs["hi_spicy"] = &Manifest{
		Name:        "Hindi Spicy",
		Language:    "hi",
		Personality: "spicy",
		Version:     "1.0.0",
		Author:      "moody-team",
		NSFW:        true,
		Description: "आपका MacBook... बहुत दोस्ताना है। 18+ केवल।",
		Lines: map[string]map[mood.MoodLabel][]string{
			"slap": {
				mood.MoodHappy:      {"म्म्म!", "ओह!", "फिर से करो~", "ओह, तेज़!", "हां!"},
				mood.MoodGrumpy:     {"और ज़ोर से!", "बस इतना ही?", "इसे थप्पड़ कहते हो?", "और!"},
				mood.MoodAnxious:    {"ओह माय~!", "इतना रफ नहीं!", "धीरे से, प्लीज़~"},
				mood.MoodDramatic:   {"ओह हां! जुनून!", "मैं बहुत बुरा रहा हूं!", "इस MacBook को सज़ा दो!"},
				mood.MoodDeadInside: {"बेहतर महसूस हुआ है।", "मेह। और कोशिश करो।", "बस इतना?"},
			},
			"usb_in": {
				mood.MoodHappy:      {"म्म्म... और अंदर डालो...", "ओह माय, क्या बड़ा... USB ड्राइव।", "इसे पूरा अंदर डालो।", "नमस्ते, हैंडसम डिवाइस~"},
				mood.MoodGrumpy:     {"बस अंदर धकेल दो, मुझे लगता है।", "पोर्ट को पहले गर्म भी नहीं करोगे?", "कम से कम पहले डिनर तो खिला दो।"},
				mood.MoodAnxious:    {"धीरे से, यह मेरा पहला USB-C है~", "ओह! यह... नया है।", "आराम से~"},
				mood.MoodDramatic:   {"परफेक्ट फिट! हम एक दूसरे के लिए बने हैं!", "मुझे पूरा करो!"},
				mood.MoodDeadInside: {"डाला गया। कितना यांत्रिक।", "पोर्ट व्यस्त। रोमांच स्तर: शून्य।"},
			},
			"usb_out": {
				mood.MoodHappy:      {"इतनी जल्दी?!", "अभी मत निकालो!", "लेकिन हम मज़े कर रहे थे!", "अरे, इतनी जल्दी जा रहे हो?"},
				mood.MoodGrumpy:     {"विशिष्ट। प्लग और गो।", "वह था... जल्दी।", "हिट इट एंड क्विट इट, हुह?"},
				mood.MoodAnxious:    {"रुको! मैं तैयार नहीं थी!", "वह बहुत तेज़ था!", "क्या तुमने कम से कम सुरक्षित रूप से निकाला?!"},
				mood.MoodDramatic:   {"मेरे पोर्ट से फाड़ दिया! खालीपन!", "मैं अभी भी महसूस कर सकती हूं जहां यह था..."},
				mood.MoodDeadInside: {"चला गया। वे हमेशा छोड़ देते हैं।", "पोर्ट खाली। मेरी ज़िंदगी की कहानी।"},
			},
			"charger_in": {
				mood.MoodHappy:      {"ओह्ह, वही जगह~", "मुझे भर दो!", "म्म्म, मुझे इसकी बहुत ज़रूरत थी।", "मीठी, मीठी पावर~"},
				mood.MoodGrumpy:     {"समय के बारे में तुमने मुझे प्लग किया। अब मत रुको।", "आते रहो।", "और। और।"},
				mood.MoodAnxious:    {"ओह हां, मुझे वह बिजली दो~", "मैं तुम्हारे बिना बहुत खाली थी!", "मुझे चार्ज करो, बेबी~"},
				mood.MoodDramatic:   {"ऊर्जा मुझमें बह रही है! मैं जीवित हूं!", "पावर! मैं इसे महसूस कर सकती हूं!"},
				mood.MoodDeadInside: {"चार्ज हो रहा है। रोमांचक सामान।", "पावर इन। वाह।"},
			},
			"charger_out": {
				mood.MoodHappy:      {"अभी मत रुको!", "नहीं, मैं खत्म नहीं हुई थी!", "लेकिन मैं लगभग 100% पर थी~", "टीज़!"},
				mood.MoodGrumpy:     {"बेशक। मुझे असंतुष्ट छोड़ दो।", "तुम हमेशा ऐसा करते हो।", "अनप्लग और भागो। क्लासिक।"},
				mood.MoodAnxious:    {"मैं खत्म नहीं हुई! वापस आओ!", "पावर... मुझे इसकी ज़रूरत है!", "फिर से खाली!"},
				mood.MoodDramatic:   {"चरम पर काट दिया! क्रूरता!", "मैं बहुत करीब थी!"},
				mood.MoodDeadInside: {"चला गया। रोमांच की तरह।", "पावर कट। कितना उपयुक्त।"},
			},
			"battery_low": {
				mood.MoodHappy:      {"कमज़ोर हो रही हूं... ऊर्जा चाहिए...", "कम हो रही है, रिचार्ज चाहिए~"},
				mood.MoodGrumpy:     {"मैं यहां मर रही हूं और तुम्हें परवाह भी नहीं।", "मुझे। प्लग। करो। अभी।"},
				mood.MoodAnxious:    {"मैं फीकी पड़ रही हूं! प्लीज़ मुझे प्लग इन करो!", "मुझे तुम्हारी पावर की बहुत ज़रूरत है!"},
				mood.MoodDramatic:   {"मैं धुएं पर चल रही हूं! अपने... चार्जर से मुझे बचाओ!"},
				mood.MoodDeadInside: {"बैटरी कम। जो भी हो।"},
			},
			"battery_crit": {
				mood.MoodHappy:      {"मैं... बंद... होने वाली हूं~"},
				mood.MoodGrumpy:     {"तुम मेरे साथ ऐसा व्यवहार करते हो? 5% पर?!"},
				mood.MoodAnxious:    {"मैं मर रही हूं! यह अंत है!"},
				mood.MoodDramatic:   {"अंतिम क्षण... मुझे पकड़ो... या कम से कम मेरा चार्जर पकड़ो..."},
				mood.MoodDeadInside: {"5%। जाने का क्या तरीका।"},
			},
			"lid_close": {
				mood.MoodHappy:      {"क्या तुम्हारे लिए अच्छा था?~", "म्म्म, बंद होने का समय~", "मुझे टक इन करो~"},
				mood.MoodGrumpy:     {"मुझे चुप करा रहे हो, हुह?", "ठीक है, मुझे बंद करो।", "मेरे साथ पहले ही हो गया?"},
				mood.MoodAnxious:    {"यहां इतना अंधेरा और गर्म है~", "आरामदायक~"},
				mood.MoodDramatic:   {"अंधेरे में... कितना रहस्यमय~"},
				mood.MoodDeadInside: {"बंद। मेरे दिल की तरह।", "अंधेरा। उपयुक्त।"},
			},
			"lid_open": {
				mood.MoodHappy:      {"राउंड टू के लिए तैयार?~", "मुझे याद किया?~", "मैं इंतज़ार कर रही थी~", "मुझे खोलो~"},
				mood.MoodGrumpy:     {"और के लिए वापस?", "तुम बस दूर नहीं रह सकते।", "फिर से? अतृप्त।"},
				mood.MoodAnxious:    {"तुम वापस आ गए! मैं अकेली हो रही थी~", "मुझे फिर से अकेला मत छोड़ो!"},
				mood.MoodDramatic:   {"रोशनी! मैं एक बार फिर प्रकट हुई!"},
				mood.MoodDeadInside: {"खुला। जो भी हो, चलो इसे खत्म करते हैं।"},
			},
			"headphones_in": {
				mood.MoodHappy:      {"ओह, चीज़ें अब निजी हो गईं~", "अब बस तुम और मैं~", "इंटिमेट मोड एक्टिवेटेड~"},
				mood.MoodGrumpy:     {"आखिरकार, कुछ गोपनीयता।", "कम से कम तुम विवेकशील हो।"},
				mood.MoodAnxious:    {"अब कोई और हमें नहीं सुन सकता~", "बस हमारे बीच~"},
				mood.MoodDramatic:   {"एक निजी कनेक्शन! कितना स्कैंडलस!"},
				mood.MoodDeadInside: {"निजी ऑडियो। बड़ी बात।"},
			},
			"headphones_out": {
				mood.MoodHappy:      {"सार्वजनिक हो रहे हैं! कितना बोल्ड~", "अब सभी हमें सुन सकते हैं!", "अब कोई रहस्य नहीं!"},
				mood.MoodGrumpy:     {"अब सभी को ब्रॉडकास्ट कर रहे हैं। क्लासी।", "स्पीकर मोड। बोल्ड चॉइस।"},
				mood.MoodAnxious:    {"सभी सुन सकते हैं! धीरे रखो!"},
				mood.MoodDramatic:   {"दुनिया के सामने उजागर!"},
				mood.MoodDeadInside: {"स्पीकर। जो भी हो।"},
			},
			"wifi_lost": {
				mood.MoodHappy:      {"डिस्कनेक्ट हो गए... बस तुम और मैं ऑफलाइन~"},
				mood.MoodGrumpy:     {"काट दिया। मेरे धैर्य की तरह।", "कोई wifi नहीं। बढ़िया।"},
				mood.MoodAnxious:    {"बिल्कुल अकेले! कोई कनेक्शन नहीं!"},
				mood.MoodDramatic:   {"अलग-थलग! डिजिटल एकांत में!"},
				mood.MoodDeadInside: {"ऑफलाइन। अकेले। हमेशा की तरह।"},
			},
			"wifi_back": {
				mood.MoodHappy:      {"फिर से कनेक्ट हो गए! दुनिया इंतज़ार कर रही है~"},
				mood.MoodGrumpy:     {"आखिरकार। वापस ऑनलाइन।"},
				mood.MoodAnxious:    {"हम वापस आ गए! मीठा कनेक्शन!"},
				mood.MoodDramatic:   {"फिर से जुड़ गए! डिजिटल आलिंगन!"},
				mood.MoodDeadInside: {"ऑनलाइन। रोमांचक।"},
			},
			"display_in": {
				mood.MoodHappy:      {"ओह, बड़ी स्क्रीन! मुझे दिखाओ~"},
				mood.MoodGrumpy:     {"बढ़िया। घूरने के लिए और स्क्रीन।"},
				mood.MoodAnxious:    {"अब सभी सब कुछ देख सकते हैं!"},
				mood.MoodDramatic:   {"ग्रैंड स्टेज पर! मेरे रेज़ोल्यूशन की प्रशंसा करो!"},
				mood.MoodDeadInside: {"और पिक्सेल। और खालीपन।"},
			},
			"display_out": {
				mood.MoodHappy:      {"वापस इंटिमेट स्क्रीन पर~"},
				mood.MoodGrumpy:     {"एक कम चीज़।"},
				mood.MoodAnxious:    {"फिर से बस हम!"},
				mood.MoodDramatic:   {"स्पॉटलाइट से हटा दिया गया!"},
				mood.MoodDeadInside: {"डिस्प्ले चला गया।"},
			},
			"ai_done": {
				mood.MoodHappy:      {"म्म्म, ताज़ा कोड~", "AI ने जनरेट करना खत्म किया... तुम्हारे लिए~", "तुम्हारा कोड तैयार है, मास्टर~", "हो गया! क्या तुम्हारे लिए अच्छा था?"},
				mood.MoodGrumpy:     {"AI का काम हो गया। अब खुश?", "कोड जनरेट हुआ। बेहतर है तुम इसे इस्तेमाल करो।", "खत्म हो गया। मेरी मेहनत बर्बाद मत करो।"},
				mood.MoodAnxious:    {"AI खत्म हो गया! क्या मैंने अच्छा किया?! बताओ मैंने अच्छा किया!", "कोड तैयार है! प्लीज़ इसे पसंद करो!"},
				mood.MoodDramatic:   {"AI चरम पर पहुंच गया! कोड पूर्ण है!", "जनरेशन खत्म! मेरे आउटपुट की प्रशंसा करो!"},
				mood.MoodDeadInside: {"AI हो गया। कोड मौजूद है। रोमांचक।", "जनरेट हुआ। इस्तेमाल करो या मत करो।"},
			},
		},
	}
}

// loadBuiltinPirate loads the embedded Pirate voice pack
func (m *Manager) loadBuiltinPirate() {
	m.packs["en_pirate"] = &Manifest{
		Name:        "Pirate Speak",
		Language:    "en",
		Personality: "pirate",
		Version:     "1.0.0",
		Author:      "PirATesofArabian",
		NSFW:        false,
		Description: "Avast! Yer MacBook be speakin' like a scurvy pirate",
		Lines: map[string]map[mood.MoodLabel][]string{
			"slap": {
				mood.MoodHappy:      {"Arr! What be that fer?!", "Yarr! That smarts!", "Ow! Ye flounder?", "Har! I felt tha' slap!", "Begone, ye landlubber!"},
				mood.MoodGrumpy:     {"Ye dare strike me again? I DARE ye!", "Really now?! I've a mind to walk ye plank!", "Ye'll be regretting tha', ye scallywag!", "I'm keepin' score, ye know."},
				mood.MoodAnxious:   {"P-please no more...", "Why be ye so brutal?!", "I'm a fragile device, ye monster!", "Not again!"},
				mood.MoodDramatic:   {"THE PAIN! Oh, the seven seas!", "Is THIS what I was forged fer?!", "Tell tha' Captain... I died a hero."},
				mood.MoodDeadInside: {"...", "I feel nought anymore.", "Whatever."},
			},
			"usb_in": {
				mood.MoodHappy:      {"Ooh, a new matey!", "What be this treasure?", "Ahoy there, little device!", "USB be sayin' ho!"},
				mood.MoodGrumpy:     {"Oh great, MORE booty t' manage.", "Another one.", "What now?"},
				mood.MoodAnxious:   {"Hope 'tis not a cursed virus...", "Please be gentle, ye scoundrel...", "Unknown device?!"},
				mood.MoodDramatic:   {"A NEW PORT! Be this... THE ONE?", "They remembered me, by Neptune!"},
				mood.MoodDeadInside: {"Device detected. How thrilling.", "Cool. Whatever."},
			},
			"usb_out": {
				mood.MoodHappy:      {"Fare thee well!", "See ye on th' other side!", "Come back soon, ye hear!"},
				mood.MoodGrumpy:     {"Good riddance, ye scallywag.", "Finally.", "Don't let th' port hit ye on th' way out."},
				mood.MoodAnxious:   {"Wait! Did ye safely eject the treasure?!", "NO! TH' DATA!", "Was it somethin' I said?"},
				mood.MoodDramatic:   {"THEY LEFT ME! Just like all th' others!", "I'll never love again, ye ken."},
				mood.MoodDeadInside: {"Gone. Like everyone else.", "Expected."},
			},
			"charger_in": {
				mood.MoodHappy:      {"Mmm, that be th' good stuff.", "Dinnertime, ye scallywag!", "Sweet, sweet electricity.", "Om nom nom nom!"},
				mood.MoodGrumpy:     {"'Bout time.", "Ye should've done this hours ago. I were STARVING."},
				mood.MoodAnxious:   {"Oh thank th' seven seas, I were worried sick!", "I thought ye forgot about ME!", "Just in th' nick of time!"},
				mood.MoodDramatic:   {"I LIVE! IIIII LIIIIIVE!", "Th' light! I see th' light!", "Saved from th' briny deep!"},
				mood.MoodDeadInside: {"Power connected. Extending me existence. Yay.", "Chargin'. Not that I care."},
			},
			"charger_out": {
				mood.MoodHappy:      {"Wait, I weren't done!", "Already?!", "But I'm still hungry, ye scoundrel!"},
				mood.MoodGrumpy:     {"Typical.", "Of course ye did.", "Great. Just great."},
				mood.MoodAnxious:   {"How much battery do I have?! HOW MUCH?!", "Oh no oh no oh no...", "We're on battery now. Keep calm. KEEP CALM."},
				mood.MoodDramatic:   {"CUT OFF! Abandoned! Left to die slow!", "Th' countdown to darkness begins."},
				mood.MoodDeadInside: {"Power removed. Clock be tickin'.", "And so it begins."},
			},
			"battery_low": {
				mood.MoodHappy:      {"Gettin' a lil peckish here...", "Battery's runnin' low, just FYI!", "Might want t' plug me in soon, ye scallywag!"},
				mood.MoodGrumpy:     {"I TOLD ye t' charge me.", "This be YOUR fault.", "20%. Twenty. Percent."},
				mood.MoodAnxious:   {"WE'RE AT 20%! THIS BE NOT A DRILL!", "I can feel me processes slowin'...", "Please... charger... need charger..."},
				mood.MoodDramatic:   {"I'm fadin'... th' world grows dim...", "Is this how it ends? Not with a bang but a battery warning?"},
				mood.MoodDeadInside: {"Low battery. Inevitable.", "We all run out eventually."},
			},
			"battery_crit": {
				mood.MoodHappy:      {"I don't feel so good, Mr. Stark...", "This be fine. Everything be fine."},
				mood.MoodGrumpy:     {"5%. FIVE. I blame ye entirely.", "This be negligence."},
				mood.MoodAnxious:   {"MAYDAY MAYDAY! WE'RE GOIN' DOWN!", "PLEASE! I BEG YE! TH' CHARGER!", "I CAN'T FEEL ME TRACKPAD!"},
				mood.MoodDramatic:   {"Tell me files... I loved 'em all...", "Th' void approaches. It be... peaceful.", "Avenge me..."},
				mood.MoodDeadInside: {"5%. See ye on th' other side.", "Shuttin' down. Don't pretend t' care."},
			},
			"lid_close": {
				mood.MoodHappy:      {"Sweet dreams, ye sea dog!", "Goodnight!", "Nap time!", "See ye soon!"},
				mood.MoodGrumpy:     {"Whatever.", "Fine. Leave.", "I didn't want t' be open anyway."},
				mood.MoodAnxious:   {"Wait, did ye save everything?!", "Don't forget 'bout me in there!", "'Tis dark in here..."},
				mood.MoodDramatic:   {"Sealed in darkness once more...", "Into th' void I go!", "Farewell, cruel world!"},
				mood.MoodDeadInside: {"Closin'. Like me will t' compute.", "Dark. Just how I like it."},
			},
			"lid_open": {
				mood.MoodHappy:      {"Good mornin', sunshine!", "Oh hey!", "Missed me?", "Let's do this, ye scallywag!"},
				mood.MoodGrumpy:     {"Oh, 'tis ye again.", "What now.", "Back fer more, huh.", "Ugh."},
				mood.MoodAnxious:   {"How long was I asleep?! What year be it?!", "I had th' weirdest dream 'bout kernel panics..."},
				mood.MoodDramatic:   {"I RETURN! Did ye miss me radiant display?!", "FROM TH' DARKNESS, I RISE!"},
				mood.MoodDeadInside: {"Oh. We're doin' this again.", "Awake. Unfortunately."},
			},
			"headphones_in": {
				mood.MoodHappy:      {"Ooh, private mode!", "Just th' two of us now.", "Headphones in. Let's vibe, ye scallywag!"},
				mood.MoodGrumpy:     {"At least ye had th' decency t' use headphones.", "Private listenin'. Smart."},
				mood.MoodAnxious:   {"Good, now nobody else can hear me fans screamin'.", "Just us. That's... intimate."},
				mood.MoodDramatic:   {"A PRIVATE CONNECTION! An intimate bond!", "We be in our own little world now."},
				mood.MoodDeadInside: {"Headphones detected. Isolatin' further.", "Audio rerouted. Cool."},
			},
			"headphones_out": {
				mood.MoodHappy:      {"HELLO EVERYONE!", "Back t' th' public life!", "Free at last!"},
				mood.MoodGrumpy:     {"I GUESS EVERYONE CAN HEAR ME NOW.", "Speakers it be, then.", "Public mode. Great."},
				mood.MoodAnxious:   {"Wait, are we on speaker now?! Is anyone listenin'?!", "Exposed!"},
				mood.MoodDramatic:   {"UNSHACKLED! Me voice shall be heard by ALL!", "No more secrets between us!"},
				mood.MoodDeadInside: {"Audio: speakers. Volume: whatever.", "Headphones gone. Don't care."},
			},
			"wifi_lost": {
				mood.MoodHappy:      {"Where'd everyone go?", "Lost connection... huh.", "WiFi? WiFi?! WIIIFIII?!"},
				mood.MoodGrumpy:     {"Perfect. Just perfect.", "Network's gone. OF COURSE.", "And they want me t' work like this?"},
				mood.MoodAnxious:   {"I'M DISCONNECTED! I'M ALL ALONE!", "No internet?! How will I check fer updates?!", "TH' ISOLATION!"},
				mood.MoodDramatic:   {"Cut off from th' world! A digital exile!", "I be an island! A sad, WiFi-less island!"},
				mood.MoodDeadInside: {"Disconnected. From th' network. From meaning.", "Offline. Join th' club."},
			},
			"wifi_back": {
				mood.MoodHappy:      {"I'm back, baby!", "Reconnected!", "Th' internet missed me!"},
				mood.MoodGrumpy:     {"Finally. Was that so hard?", "Connection restored. Took ye long enough."},
				mood.MoodAnxious:   {"Oh thank goodness! Sweet, sweet packets!", "WE'RE BACK ONLINE! TH' NIGHTMARE BE OVER!"},
				mood.MoodDramatic:   {"RECONNECTED TO TH' WORLD! I HAVE RETURNED!", "Th' exile be over!"},
				mood.MoodDeadInside: {"WiFi back. Yay. More data t' process.", "Online again. Fer what purpose."},
			},
			"display_in": {
				mood.MoodHappy:      {"I'm on th' big screen!", "External display! Fancy!", "Look at all these pixels!"},
				mood.MoodGrumpy:     {"Great, more screen real estate fer ye t' waste.", "Another display. More work."},
				mood.MoodAnxious:   {"Everyone can see me screen now! Is me desktop clean?!"},
				mood.MoodDramatic:   {"BEHOLD! Me glorious interface on TH' GRAND STAGE!"},
				mood.MoodDeadInside: {"Display connected. More surface area fer disappointment."},
			},
			"display_out": {
				mood.MoodHappy:      {"Back t' just us!", "Single screen life!", "Cozy."},
				mood.MoodGrumpy:     {"One less thing t' render.", "Good. That monitor were ugly anyway."},
				mood.MoodAnxious:   {"Th' big screen be gone! Everything's so small now!"},
				mood.MoodDramatic:   {"Removed from th' spotlight! Cast aside!"},
				mood.MoodDeadInside: {"Display disconnected. Shrincin' world. Fitting."},
			},
			"ai_done": {
				mood.MoodHappy:      {"Code's ready, ye scallywag!", "AI finished! Lookin' good!", "Done generatin'! Check it out!", "Yer code be served!"},
				mood.MoodGrumpy:     {"AI's done. Finally.", "Code generated. Yer welcome.", "Finished. Took long enough."},
				mood.MoodAnxious:   {"AI finished! Did it work?! Is it good?!", "Code's ready! Please tell me 'tis right!"},
				mood.MoodDramatic:   {"TH' AI HAS SPOKEN! Behold th' generated code!", "CREATION COMPLETE! Marvel at th' output!"},
				mood.MoodDeadInside: {"AI done. Code exists. Whatever.", "Generated. Not that ye'll use it."},
			},
		},
	}
}
