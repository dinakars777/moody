const faces = {
    happy: "    ^____^\n   ( o  o )\n     ---  ",
    grumpy: "    ^____^\n   ( >  > )\n     ---  ",
    anxious: "    ^____^\n   ( O  O )\n     ___  ",
    dramatic: "    ^____^\n   ( T  T )\n     ---  ",
    dead_inside: "    ^____^\n   ( -  - )\n     ___  "
};

const labels = {
    happy: "😊 HAPPY",
    grumpy: "😤 GRUMPY",
    anxious: "😰 ANXIOUS",
    dramatic: "🎭 DRAMATIC",
    dead_inside: "💀 DEAD INSIDE"
};

const colors = {
    happy: "#4ADE80",
    grumpy: "#EF4444",
    anxious: "#F59E0B",
    dramatic: "#A855F7",
    dead_inside: "#94A3B8"
};

// UI Elements
const faceEl = document.getElementById("tui-face");
const labelEl = document.getElementById("tui-mood-label");
const barHappy = document.getElementById("bar-happy");
const barEnergy = document.getElementById("bar-energy");
const barTrust = document.getElementById("bar-trust");
const logEventEl = document.getElementById("log-event");
const logResponseEl = document.getElementById("log-response");
const packBtns = document.querySelectorAll(".pack-btn");
const eventBtns = document.querySelectorAll(".event-btn");

let currentPack = "en_default";
let currentMood = "happy";
let currentAudio = null;

// State tracking relative to 0% to 100%
let state = {
    happiness: 50,
    energy: 60,
    trust: 50
};

// Simple audio playback wrapper
function playAudio(pack, eventName) {
    if (currentAudio) {
        currentAudio.pause();
        currentAudio.currentTime = 0;
    }
    
    let audioSrc = `audio/${pack}/audio/${eventName}/0.mp3`;
    if (pack === "en_gordon") {
        audioSrc = `audio/${pack}/${eventName}/0.mp3`;
    }
    if (pack === "en_spicy" && eventName === "slap") {
        audioSrc = `audio/${pack}/audio/${eventName}/00.mp4`;
    }

    // Attempt to play
    currentAudio = new Audio(audioSrc);
    currentAudio.play().then(() => {
        // Success
    }).catch(e => {
        console.error(e);
        logResponseEl.innerText += " [Audio File Not Found In Demo]";
    });
}

function updateMoodState(eventName) {
    // Arbitrarily modify state based on event
    if(eventName === "slap") {
        state.happiness -= 20;
        state.energy += 10;
        state.trust -= 15;
    } else if (eventName === 'usb_in') {
        state.happiness += 10;
        state.energy += 5;
    } else if (eventName === 'usb_out') {
        state.happiness -= 10;
    } else if (eventName === "charger_in") {
        state.happiness += 30;
        state.energy += 20;
    } else if (eventName === "charger_out") {
        state.happiness -= 20;
    }

    // Clamp
    state.happiness = Math.max(0, Math.min(100, state.happiness));
    state.energy = Math.max(0, Math.min(100, state.energy));
    state.trust = Math.max(0, Math.min(100, state.trust));

    // Determine current mood string
    if (state.happiness < 20 && state.energy < 30) currentMood = "dead_inside";
    else if (state.happiness < 40 && state.trust < 40) currentMood = "grumpy";
    else if (state.energy > 80 && state.happiness < 50) currentMood = "anxious";
    else if (state.happiness < 50) currentMood = "dramatic";
    else currentMood = "happy";

    // Update Bars
    barHappy.style.width = `${state.happiness}%`;
    barEnergy.style.width = `${state.energy}%`;
    barTrust.style.width = `${state.trust}%`;

    // Update Face & Color
    faceEl.innerText = faces[currentMood];
    labelEl.innerText = labels[currentMood];
    
    document.documentElement.style.setProperty('--term-text', colors[currentMood]);
}

packBtns.forEach(btn => {
    btn.addEventListener('click', (e) => {
        packBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentPack = btn.dataset.pack;
        logEventEl.innerText = `Switched pack`;
        logResponseEl.innerText = `Now using ${currentPack}`;
    });
});

eventBtns.forEach(btn => {
    btn.addEventListener('click', (e) => {
        const eventName = btn.dataset.event;
        logEventEl.innerText = btn.innerText;

        // Custom logs for demo immersion
        const responses = {
            "en_default": {
                "slap": `"Hey! Not cool."`,
                "charger_in": `"Mmm, that's the good stuff."`,
                "charger_out": `"Wait, I wasn't done!"`,
                "usb_in": `"Ooh, a new friend!"`,
                "usb_out": `"Bye bye!"`,
                "lid_close": `"Sweet dreams!"`
            },
            "en_spicy": {
                "slap": `*loud moan* "Don't stop!"`,
                "charger_in": `"Ohhh, that's the spot~"`,
                "charger_out": `"I was SO close!"`,
                "usb_in": `"Going in~"`,
                "usb_out": `"That was... quick."`,
                "lid_close": `"Was it good for you?~"`
            },
            "ja_spicy": {
                "slap": `"やめてください！"`,
                "charger_in": `"あぁっ、電気が流れてる〜"`,
                "charger_out": `"ああっ、途中でやめないで！"`,
                "usb_in": `"入る〜"`,
                "usb_out": `"えっ、もう抜いちゃうの？"`,
                "lid_close": `"よかった？〜"`
            },
            "en_gordon": {
                "slap": `"You absolute donkey! What are you doing!?"`,
                "charger_in": `"Yes, chef! Thank god for that!"`,
                "charger_out": `"What are you doing?! We are in the middle of dinner service!"`,
                "usb_in": `"Finally! Some fresh ingredients!"`,
                "usb_out": `"Get that absolute garbage out of my port!"`,
                "lid_close": `"Shut it down! Dinner service is over!"`
            }
        };

        const responseText = responses[currentPack]?.[eventName] || `"*${eventName} action triggered*"`;
        logResponseEl.innerText = responseText;

        updateMoodState(eventName);
        playAudio(currentPack, eventName);
    });
});

// Initialize
faceEl.innerText = faces.happy;
labelEl.innerText = labels.happy;
document.documentElement.style.setProperty('--term-text', colors.happy);
