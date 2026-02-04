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

    // --- 1. LOGIKA SCROLL HINT (Tetap Pakai ini) ---
    const scrollHint = document.querySelector('.scroll-hint');
    const hookSection = document.getElementById('the-hook');

    if (scrollHint && hookSection) {
        scrollHint.addEventListener('click', () => {
            hookSection.scrollIntoView({ behavior: 'smooth' });
        });
    }

    // --- HERO PARALLAX EFFECT ---
    const heroSection = document.querySelector('.hero-section');
    const heroContent = document.querySelector('.hero-section .content');

    if (heroSection) {
        window.addEventListener('scroll', () => {
            const scrollY = window.scrollY;
            
            // Optimization: Only animate when hero is visible
            if (scrollY <= heroSection.offsetHeight) {
                // Text moves slightly slower (0.4x) & fades out
                if (heroContent) {
                    heroContent.style.transform = `translateY(${scrollY * 0.4}px)`;
                    heroContent.style.opacity = 1 - (scrollY / 700);
                }

                // Scroll hint fades out fast
                if (scrollHint) {
                    scrollHint.style.opacity = 1 - (scrollY / 300);
                    // Keep the X centering while moving down
                    scrollHint.style.transform = `translate(-50%, ${scrollY * 0.5}px)`;
                }
            }
        });
    }

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