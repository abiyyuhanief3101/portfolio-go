const CACHE_NAME = 'abiyyu-portfolio-v1';
const STATIC_ASSETS = [
    '/',
    '/css/style.css',
    '/js/script.js',
    '/manifest.json'
];

// Install Service Worker & Cache aset awal
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => {
            return cache.addAll(STATIC_ASSETS);
        })
    );
});

// Intercept requests: Strategi Network-First dengan Fallback ke Cache
self.addEventListener('fetch', (event) => {
    event.respondWith(
        fetch(event.request)
            .then((response) => {
                // Simpan ke cache jika sukses untuk dibaca saat offline nanti
                if (response.status === 200 && event.request.method === 'GET') {
                    const responseClone = response.clone();
                    caches.open(CACHE_NAME).then((cache) => {
                        cache.put(event.request, responseClone);
                    });
                }
                return response;
            })
            .catch(() => caches.match(event.request)) // Fallback mode saat offline
    );
});