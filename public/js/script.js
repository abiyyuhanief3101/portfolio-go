document.addEventListener('DOMContentLoaded', () => {
    
    // --- THEME TOGGLE LOGIC ---
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            const root = document.documentElement;
            const isDark = root.getAttribute('data-theme') === 'dark';
            
            if (isDark) {
                root.removeAttribute('data-theme');
                localStorage.setItem('theme', 'light');
            } else {
                root.setAttribute('data-theme', 'dark');
                localStorage.setItem('theme', 'dark');
            }
        });
    }

    // Ambil elemen modal
    const modal = document.getElementById('modal');
    const closeBtn = document.querySelector('.close-btn');
    const closeActionBtn = document.querySelector('.close-action');
    
    // Ambil semua tombol yang punya kelas 'coming-soon'
    const triggers = document.querySelectorAll('.coming-soon');

    // Fungsi buka modal
    function openModal(e) {
        e.preventDefault(); // Mencegah link pindah halaman
        modal.style.display = 'flex';
    }

    // Fungsi tutup modal
    function closeModal() {
        modal.style.display = 'none';
    }

    // Pasang event listener ke setiap tombol pemicu
    triggers.forEach(btn => {
        btn.addEventListener('click', openModal);
    });

    // Pasang event listener ke tombol tutup
    closeBtn.addEventListener('click', closeModal);
    closeActionBtn.addEventListener('click', closeModal);

    // Tutup kalau user klik di luar kotak modal (di area gelap)
    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            closeModal();
        }
    });

    console.log("Humanist System Ready! 🌿");

    // --- TYPING EFFECT (Badge / Hello World) ---
    const badge = document.querySelector('.badge');
    if (badge) {
        // Daftar kata yang akan berganti-ganti
        const phrases = ["👋 Hello World!", "Thank you for visiting", "Let's Collaborate"];
        let phraseIndex = 0;
        let charIndex = 0;
        let isDeleting = false;
        
        function typeWriter() {
            const currentPhrase = phrases[phraseIndex];
            
            if (isDeleting) {
                badge.textContent = currentPhrase.substring(0, charIndex - 1);
                charIndex--;
            } else {
                badge.textContent = currentPhrase.substring(0, charIndex + 1);
                charIndex++;
            }
            
            let typeSpeed = isDeleting ? 50 : 100;

            if (!isDeleting && charIndex === currentPhrase.length) {
                typeSpeed = 5000; // Tunggu 2 detik sebelum menghapus
                isDeleting = true;
            } else if (isDeleting && charIndex === 0) {
                isDeleting = false;
                phraseIndex = (phraseIndex + 1) % phrases.length;
                typeSpeed = 700; // Tunggu sebentar sebelum mengetik kata baru
            }

            setTimeout(typeWriter, typeSpeed);
        }
        typeWriter();
    }

    // --- 1. Hook Section Reference ---
    const hookSection = document.getElementById('the-hook');

    // --- HERO PARALLAX EFFECT ---
    const heroSection = document.querySelector('.hero-section');
    const heroContent = document.querySelector('.hero-section .content');

    

    // --- LOGOS: Fade In & Parallax Effect ---
    const logos = document.querySelectorAll('.float-logo');

    if (hookSection && logos.length > 0) {
        // 1. Fade In Effect (Intersection Observer)
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    logos.forEach((logo, index) => {
                        setTimeout(() => {
                            logo.classList.add('visible');
                        }, index * 300); // Muncul satu per satu (staggered)
                    });
                    observer.unobserve(entry.target);
                }
            });
        }, { threshold: 0.3 });

        observer.observe(hookSection);

        // 2. Parallax Effect (Scroll)
        window.addEventListener('scroll', () => {
            const sectionTop = hookSection.offsetTop;
            const scrollY = window.scrollY;
            
            // Hitung jarak scroll relatif terhadap section
            const relativeScroll = scrollY - sectionTop;
            
            logos.forEach((logo, index) => {
                // Kecepatan beda-beda biar kerasa 3D
                // Kita pakai margin-top supaya tidak menimpa animasi CSS transform (float)
                const speed = (index % 2 === 0 ? 0.05 : -0.03) * (index + 1); 
                logo.style.marginTop = `${relativeScroll * speed}px`;
            });
        });
    }

    // --- PLAYFUL FEATURES LOGIC ---
    
    // 1. Elements Selection
    const avatarContainer = document.getElementById('guide-avatar');
    const avatarImg = avatarContainer ? avatarContainer.querySelector('img') : null;
    const avatarTooltip = avatarContainer ? avatarContainer.querySelector('.avatar-tooltip') : null;
    const timelineLinks = document.querySelectorAll('.timeline-nav a');
    const sections = document.querySelectorAll('section');

    // 2. Intersection Observer for Active State
    const observerOptions = {
        threshold: 0.5 // Trigger when 50% of the section is visible
    };

    const observerCallback = (entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const id = entry.target.getAttribute('id');

                // A. Update Timeline
                timelineLinks.forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('href') === `#${id}`) {
                        link.classList.add('active');
                    }
                });

                // B. Update Avatar (Chameleon Effect)
                if (avatarImg && avatarTooltip) {
                    let newSrc = '/img/avatar_default.png';
                    let message = "";

                    if (id === 'the-paradox' || id === 'journey') {
                        newSrc = '/img/avatar_nature.png';
                        message = "Back to Roots! 🌱";
                    } else if (id === 'ascent') {
                        newSrc = '/img/avatar_tech.png';
                        message = "Tech Mode: ON 🚀";
                    } else {
                        newSrc = '/img/avatar_default.png';
                        message = "Hi there! 👋";
                    }

                    // Only update if src changes to avoid flickering
                    if (!avatarImg.src.includes(newSrc)) {
                        avatarImg.style.opacity = 0;
                        setTimeout(() => {
                            avatarImg.src = newSrc;
                            avatarImg.style.opacity = 1;
                            
                            // Pop the tooltip
                            avatarTooltip.textContent = message;
                            avatarTooltip.classList.add('show');
                            setTimeout(() => {
                                avatarTooltip.classList.remove('show');
                            }, 3000);
                        }, 300);
                    }
                }
            }
        });
    };

    // Start Observing
    if (sections.length > 0) {
        const sectionObserver = new IntersectionObserver(observerCallback, observerOptions);
        sections.forEach(section => sectionObserver.observe(section));
    }

    // --- SCROLL REVEAL ANIMATION ---
    const revealElements = document.querySelectorAll('.reveal');

    const revealObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('active');
                revealObserver.unobserve(entry.target); // Only animate once
            }
        });
    }, {
        threshold: 0.40, // Trigger when 15% of element is visible
        rootMargin: "0px 0px -50px 0px" // Trigger slightly before it hits the bottom
    });

    revealElements.forEach(el => revealObserver.observe(el));
    
    console.log("Humanist System Ready! 🌿");
    // --- EASTER EGG: CONFETTI RECENT WIN ---
    const confettiBtn = document.getElementById('confetti-trigger');
    if (confettiBtn) {
        confettiBtn.addEventListener('click', () => {
            // Animasi Confetti dasar
            confetti({
                particleCount: 100,
                spread: 70,
                origin: { y: 0.6 },
                colors: ['#3F756C', '#C6743E', '#F6EECD'] // Pakai warna brand kamu
            });
            
            // Ganti teks sementara
            const originalText = confettiBtn.innerText;
            confettiBtn.innerText = "🚀 Automaton Deployed!";
            setTimeout(() => {
                confettiBtn.innerText = originalText;
            }, 2000);
        });
    }
   
});