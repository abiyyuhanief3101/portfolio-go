/* Site-wide behaviour. Everything else on this site works without JavaScript. */
(function () {
    var toggle = document.querySelector('.nav-toggle');
    var list = document.getElementById('nav-list');
    if (!toggle || !list) return;

    function setOpen(open) {
        list.classList.toggle('is-open', open);
        toggle.setAttribute('aria-expanded', String(open));
    }

    toggle.addEventListener('click', function () {
        setOpen(toggle.getAttribute('aria-expanded') !== 'true');
    });

    list.addEventListener('click', function (e) {
        if (e.target.closest('a')) setOpen(false);
    });

    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape' && toggle.getAttribute('aria-expanded') === 'true') {
            setOpen(false);
            toggle.focus();
        }
    });

    document.addEventListener('click', function (e) {
        if (!e.target.closest('.site-nav') && toggle.getAttribute('aria-expanded') === 'true') {
            setOpen(false);
        }
    });
})();

/* Service worker: network-first with an offline fallback cache. */
if ('serviceWorker' in navigator) {
    window.addEventListener('load', function () {
        navigator.serviceWorker.register('/sw.js').catch(function () { /* offline support is optional */ });
    });
}
