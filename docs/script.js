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
const packSelectorEl = document.getElementById("pack-selector");
const eventBtns = document.querySelectorAll(".event-btn");

let currentPack = "en_spicy";
let currentMood = "happy";
let currentAudio = null;
let demoConfig = null;
let packsById = new Map();

// State tracking relative to 0% to 100%
let state = {
    happiness: 50,
    energy: 60,
    trust: 50
};

// Simple audio playback wrapper
function playAudio(pack, eventName) {
    const packConfig = packsById.get(pack);
    if (!packConfig || packConfig.textOnly || !packConfig.audio) {
        return;
    }
    
    if (currentAudio) {
        currentAudio.pause();
        currentAudio.currentTime = 0;
    }
    
    const eventOverrides = packConfig.audio.events || {};
    const extension = packConfig.audio.extension || "mp3";
    const audioSrc = eventOverrides[eventName] || `${packConfig.audio.base}/${eventName}/0.${extension}`;

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

function renderPackSelector(config) {
    packSelectorEl.innerHTML = "";
    packsById = new Map(config.packs.map(pack => [pack.id, pack]));

    config.packs.forEach(pack => {
        const btn = document.createElement("button");
        btn.className = "pack-btn";
        btn.dataset.pack = pack.id;
        btn.type = "button";
        btn.innerText = pack.label;
        if (pack.id === currentPack) {
            btn.classList.add("active");
        }

        btn.addEventListener('click', () => {
            document.querySelectorAll(".pack-btn").forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentPack = btn.dataset.pack;
            logEventEl.innerText = `Switched pack`;
            logResponseEl.innerText = `Now using ${currentPack}`;
        });

        packSelectorEl.appendChild(btn);
    });
}

async function loadDemoConfig() {
    const response = await fetch("packs.json");
    if (!response.ok) {
        throw new Error(`Failed to load packs.json: ${response.status}`);
    }
    return response.json();
}

eventBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        if (!demoConfig) {
            logResponseEl.innerText = "Demo configuration is still loading.";
            return;
        }
        const eventName = btn.dataset.event;
        const packConfig = packsById.get(currentPack);
        logEventEl.innerText = btn.innerText;

        const responseText = packConfig?.responses?.[eventName] || `"*${eventName} action triggered*"`;
        logResponseEl.innerText = responseText;

        updateMoodState(eventName);
        playAudio(currentPack, eventName);
    });
});

async function initializeDemo() {
    try {
        demoConfig = await loadDemoConfig();
        currentPack = demoConfig.defaultPack || demoConfig.packs[0]?.id || currentPack;
        renderPackSelector(demoConfig);
    } catch (error) {
        console.error(error);
        packSelectorEl.innerHTML = `<button class="pack-btn active" disabled>Demo config unavailable</button>`;
        logResponseEl.innerText = "Demo config could not be loaded.";
    }

    faceEl.innerText = faces.happy;
    labelEl.innerText = labels.happy;
    document.documentElement.style.setProperty('--term-text', colors.happy);
}

initializeDemo();
