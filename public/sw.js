const CACHE_NAME = 'abiyyu-portfolio-v2';
const STATIC_ASSETS = [
    '/',
    '/css/style.css',
    '/js/script.js',
    '/manifest.json'
];

// Install Service Worker & Cache aset awal
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
    );
});

// Buang cache versi lama supaya aset desain lama tidak tertinggal
self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((keys) => Promise.all(
            keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))
        ))
    );
});

// Intercept requests: Strategi Network-First dengan Fallback ke Cache
self.addEventListener('fetch', (event) => {
    event.respondWith(
        fetch(event.request)
            .then((response) => {
                // Simpan ke cache jika sukses untuk dibaca saat offline nanti
                if (response.status === 200 && event.request.method === 'GET' && event.request.url.startsWith(self.location.origin)) {
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
