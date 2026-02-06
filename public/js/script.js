document.addEventListener('DOMContentLoaded', () => {
    
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
                typeSpeed = 2000; // Tunggu 2 detik sebelum menghapus
                isDeleting = true;
            } else if (isDeleting && charIndex === 0) {
                isDeleting = false;
                phraseIndex = (phraseIndex + 1) % phrases.length;
                typeSpeed = 500; // Tunggu sebentar sebelum mengetik kata baru
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
    
    console.log("Humanist System Ready! 🌿");
   
});