// ==========================================
// 1. STATE & DATA MANAGEMENT
// ==========================================
const supabaseUrl = 'https://kgscotrveqoixnufzxea.supabase.co';
const supabaseKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imtnc2NvdHJ2ZXFvaXhudWZ6eGVhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzUzMDA3MjIsImV4cCI6MjA5MDg3NjcyMn0.2cjGOOcuyxE1z-5yhQo1epzfFd92nGPBDgPshCTbBi8';
const supabaseClient = supabase.createClient(supabaseUrl, supabaseKey);

let currentState = { screen: 0, language: 'en', questionIndex: 0, answers: [] };

const i18n = {
    id: {
        title: "Discover Your 3 Layers.",
        desc: "Eksplorasi psikologi singkat (60 detik) untuk melihat bagaimana kamu memandang dirimu sendiri dan dunia.",
        discTitle: "💡 Catatan:",
        discText: "Ini adalah icebreaker psikologi populer yang didasarkan pada Efek Barnum. Dirancang murni untuk seru-seruan dan refleksi diri, bukan diagnosis klinis.",
        startBtn: "Start the Journey 🚀",
        labelAnimal: "Hewan apa itu?",
        labelReason: "Satu alasan utama kenapa kamu menyukainya?",
        btnNext: "Lanjut",
        questions: [
            "Mari kita mulai. Coba sebutkan hewan favorit pertamamu (No. 1). Hewan apa itu?",
            "Oke, sekarang sebutkan hewan favorit keduamu (No. 2). Hewan apa ini?",
            "Terakhir, sebutkan hewan favorit ketigamu (No. 3). Pastikan berbeda dari yang sebelumnya ya. Hewan apa itu?"
        ],
        loadingTxt: "Menganalisis pola pikirmu...",
        formTitle: "Analisis Selesai! 🧠",
        formDesc: "Kami menemukan pola menarik dari jawabanmu. Masukkan nama dan profesimu. (Masukkan emailmu jika ingin kami mengirimkan salinan hasilnya!).",
        lName: "Nama / Panggilan",
        lRole: "Profesi / Role",
        lEmail: "Email (Opsional)",
        btnReveal: "Reveal My Profile 🔓"
    },
    en: {
        title: "Discover Your 3 Layers.",
        desc: "A 60-second psychological exploration to see how you truly view yourself and the world.",
        discTitle: "💡 Disclaimer:",
        discText: "This is a popular psychological icebreaker based on the Barnum Effect. It's designed purely for fun and introspection, not clinical diagnosis.",
        startBtn: "Start the Journey 🚀",
        labelAnimal: "What animal is it?",
        labelReason: "Give one main reason why you like it.",
        btnNext: "Next",
        questions: [
            "Let's start. Name your first favorite animal (No. 1). What is it?",
            "Great, now name your second favorite animal (No. 2). What is it?",
            "Lastly, name your third favorite animal (No. 3). Make sure it's different from the previous two. What is it?"
        ],
        loadingTxt: "Analyzing your thought patterns...",
        formTitle: "Analysis Complete! 🧠",
        formDesc: "We found interesting patterns in your answers. Enter your details below. (Include your email if you'd like a copy of the results sent to you!).",
        lName: "Name / Nickname",
        lRole: "Profession / Role",
        lEmail: "Email (Optional)",
        btnReveal: "Reveal My Profile 🔓"
    }
};

// ==========================================
// 2. TEXT HELPERS
// ==========================================
function capitalizeTitle(str) {
    if (!str) return "";
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function cleanReason(str) {
    if (!str) return "";
    let cleaned = str.trim();
    cleaned = cleaned.replace(/^(karena|karna|because|cause|cuz)\s+/i, '');
    
    if (cleaned === cleaned.toUpperCase()) {
        return cleaned;
    }
    return cleaned.charAt(0).toLowerCase() + cleaned.slice(1);
}

// ==========================================
// 3. UI CONTROLLERS
// ==========================================
function setLanguage(lang) {
    currentState.language = lang;
    const t = i18n[lang];

    document.querySelectorAll('.lang-btn').forEach(btn => btn.classList.remove('active'));
    if(lang === 'en') document.querySelectorAll('.lang-btn')[0].classList.add('active');
    if(lang === 'id') document.querySelectorAll('.lang-btn')[1].classList.add('active');

    document.getElementById('t-title').innerText = t.title;
    document.getElementById('t-desc').innerText = t.desc;
    document.getElementById('t-disclaimer-title').innerText = t.discTitle;
    document.getElementById('t-disclaimer-text').innerText = t.discText;
    document.getElementById('t-start-btn').innerText = t.startBtn;

    document.getElementById('l-animal').innerText = t.labelAnimal;
    document.getElementById('l-reason').innerText = t.labelReason;
    document.getElementById('btn-next-question').innerText = t.btnNext;
    
    if(currentState.screen === 1) {
        document.getElementById('q-text').innerText = t.questions[currentState.questionIndex];
    }

    document.getElementById('t-loading').innerText = t.loadingTxt;
    document.getElementById('t-form-title').innerText = t.formTitle;
    document.getElementById('t-form-desc').innerText = t.formDesc;
    document.getElementById('l-name').innerText = t.lName;
    document.getElementById('l-role').innerText = t.lRole;
    document.getElementById('l-email').innerText = t.lEmail;
    document.getElementById('btn-reveal').innerText = t.btnReveal;
}

setLanguage(currentState.language);

function nextScreen() {
    currentState.screen = 1;
    document.getElementById('screen-0').classList.remove('active');
    document.getElementById('screen-question').classList.add('active');
    renderQuestion();
}

function renderQuestion() {
    const t = i18n[currentState.language];
    const idx = currentState.questionIndex;

    document.getElementById('q-text').innerText = t.questions[idx];
    
    const progressPercent = ((idx + 1) / 3) * 100;
    document.getElementById('progress-bar').style.width = progressPercent + '%';
    document.getElementById('progress-text').innerText = `${idx + 1} / 3`;

    const animalInput = document.getElementById('input-animal');
    const reasonInput = document.getElementById('input-reason');
    animalInput.value = '';
    reasonInput.value = '';
    
    validateInput(); 
    animalInput.addEventListener('input', validateInput);
    reasonInput.addEventListener('input', validateInput);
}

function validateInput() {
    const animal = document.getElementById('input-animal').value.trim();
    const reason = document.getElementById('input-reason').value.trim();
    const btnNext = document.getElementById('btn-next-question');
    btnNext.disabled = !(animal !== '' && reason !== '');
}

function handleQuestionSubmit() {
    const animal = document.getElementById('input-animal').value.trim();
    const reason = document.getElementById('input-reason').value.trim();

    currentState.answers.push({ animal: animal, reason: reason });

    if (currentState.questionIndex < 2) {
        currentState.questionIndex++;
        renderQuestion();
    } else {
        currentState.screen = 4;
        document.getElementById('screen-question').classList.remove('active');
        document.getElementById('screen-4').classList.add('active');
        
        setTimeout(() => {
            document.getElementById('loading-section').style.display = 'none';
            document.getElementById('form-section').style.display = 'block';
            document.getElementById('form-section').style.animation = 'fadeIn 0.5s ease forwards';
        }, 2000);
    }
}

async function revealResults() {
    const name = document.getElementById('input-name').value.trim();
    const role = document.getElementById('input-role').value.trim();
    const email = document.getElementById('input-email').value.trim();
    
    if(name === '' || role === '') {
        alert(currentState.language === 'id' ? "Mohon isi Nama dan Profesimu ya!" : "Please fill in your Name and Role!");
        return;
    }

    const btnReveal = document.getElementById('btn-reveal');
    btnReveal.innerText = "Processing...";
    btnReveal.disabled = true;

    const a1 = currentState.answers[0];
    const a2 = currentState.answers[1];
    const a3 = currentState.answers[2];

    // Kirim data ke Supabase
    try {
        const { error } = await supabaseClient.from('visitor_insights').insert([{
            name: name, role: role, email: email,
            animal_1: a1.animal, reason_1: a1.reason,
            animal_2: a2.animal, reason_2: a2.reason,
            animal_3: a3.animal, reason_3: a3.reason
        }]);
        if (error) console.error("Gagal mengirim data ke Supabase:", error);
    } catch (err) {
        console.error("Terjadi kesalahan jaringan:", err);
    }

    // Tampilkan Hasil di Layar
    currentState.screen = 5;
    document.getElementById('screen-4').classList.remove('active');
    document.getElementById('screen-5').classList.add('active');
    document.getElementById('spa-container').classList.add('wide-card');

    document.getElementById('res-animal-1').innerText = capitalizeTitle(a1.animal);
    document.getElementById('res-animal-2').innerText = capitalizeTitle(a2.animal);
    document.getElementById('res-animal-3').innerText = capitalizeTitle(a3.animal);

    const r1 = cleanReason(a1.reason);
    const r2 = cleanReason(a2.reason);
    const r3 = cleanReason(a3.reason);

    if(currentState.language === 'id') {
        document.getElementById('t-result-title').innerText = "Your 3 Layers";
        document.getElementById('res-desc-1').innerHTML = `Secara tidak sadar, kamu memproyeksikan energimu layaknya ${a1.animal}. Apresiasimu terhadap konsep <strong>"${r1}"</strong> secara metaforis mencerminkan karakter yang paling ingin kamu tunjukkan kepada dunia saat ini.`;
        document.getElementById('res-desc-2').innerHTML = `Tanpa disadari, orang lain justru paling sering menangkap auramu layaknya ${a2.animal}. Kesan yang menyiratkan <strong>"${r2}"</strong> adalah energi dominan yang mereka rasakan darimu.`;
        document.getElementById('res-desc-3').innerHTML = `Namun jauh di lubuk hatimu, jati dirimu (Core Self) beresonansi dengan ${a3.animal}. Filosofi dari <strong>"${r3}"</strong> menyiratkan nilai paling sejati yang murni ada di dalam dirimu.`;
        document.getElementById('t-closing').innerHTML = `Terkadang kita butuh melihat sesuatu dari 3 perspektif berbeda. Sama seperti bagaimana sebuah sistem dibangun: dari sisi pengguna, bisnis, dan teknis. <strong>Let's connect!</strong>`;
    } else {
        document.getElementById('t-result-title').innerText = "Your 3 Layers";
        document.getElementById('res-desc-1').innerHTML = `You unconsciously project the energy of a ${a1.animal}. Your appreciation for the concept of <strong>"${r1}"</strong> metaphorically reflects the exact identity you want the world to see right now.`;
        document.getElementById('res-desc-2').innerHTML = `Without realizing it, people often perceive you with the vibe of a ${a2.animal}. The impression that implies <strong>"${r2}"</strong> is the most prominent energy they feel from you.`;
        document.getElementById('res-desc-3').innerHTML = `Deep down, your truest core resonates with the ${a3.animal}. The philosophy behind <strong>"${r3}"</strong> implies the pure value you hold within.`;
        document.getElementById('t-closing').innerHTML = `Sometimes we need to see things from 3 different perspectives. Just like how systems are built: from the user's side, business side, and technical side. <strong>Let's connect!</strong>`;
    }

    if (email !== '') {
        const emailNotice = currentState.language === 'id' ?
            `<br><br><span class="spam-notice" style="font-size: 0.85em; color: var(--terracotta);"><em>*Hasil analisismu sedang dikirim. Cek folder <strong>Promotions</strong> atau <strong>Spam</strong> jika email belum masuk di Inbox utama.</em></span>` :
            `<br><br><span class="spam-notice" style="font-size: 0.85em; color: var(--terracotta);"><em>*Your results have been sent. Please check your <strong>Promotions</strong> or <strong>Spam</strong> folder if you don't see it in your primary inbox.</em></span>`;
        document.getElementById('t-closing').innerHTML += emailNotice;
    }

    // --- PROSES KIRIM EMAIL (HANYA TEKS) ---
    if (email !== '') {
        const emailPayload = {
            name: name,
            email: email,
            desc_1: document.getElementById('res-desc-1').innerText,
            desc_2: document.getElementById('res-desc-2').innerText,
            desc_3: document.getElementById('res-desc-3').innerText
        };

        fetch('/send-email', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(emailPayload)
        }).catch(err => console.error("Gagal mengirim trigger email:", err));
    }
}

function shareResult() {
    const shareText = currentState.language === 'en' ? 
        "I just discovered my 3 Layers of personality! Try this fun psychological exploration:" : 
        "Saya baru saja membongkar 3 Layer kepribadian saya! Coba eksplorasi seru ini:";
        
    if (navigator.share) {
        navigator.share({
            title: 'My 3 Layers',
            text: shareText,
            url: window.location.href
        }).catch(console.error);
    } else {
        navigator.clipboard.writeText(window.location.href);
        alert(currentState.language === 'en' ? "Link copied to clipboard!" : "Link disalin ke clipboard!");
    }
}