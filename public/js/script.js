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

   
});