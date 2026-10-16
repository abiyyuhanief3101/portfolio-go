// Haze: #D5D5D5 smoke clouds drifting past on the wind, some thick and
// billowing, some thin and stretched, plus a faint fire glow low on the
// screen and embers blowing across. Drawn behind the page content, so the
// page keeps its own colours and the smoke only shows in the gaps.
// With reduced motion it paints one still frame without embers.
(function () {
  var ORANGE = '247,147,1', RED = '205,46,14', ASH = [213, 213, 213];
  var WIND = 26; // px/s, left to right
  // Smoke is composited at this opacity at most, so muted body text keeps
  // 4.5:1 even behind the thickest part of a cloud.
  var SMOKE_MAX = 0.5;

  var reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
  var canvas = document.createElement('canvas');
  canvas.className = 'haze-canvas';
  canvas.setAttribute('aria-hidden', 'true');
  document.body.prepend(canvas);
  var ctx = canvas.getContext('2d');

  // Clouds are drawn into a half-resolution layer first, then laid on the
  // page in one pass. Overlapping puffs build up inside the layer, but the
  // final opacity never passes SMOKE_MAX.
  var smoke = document.createElement('canvas');
  var sctx = smoke.getContext('2d');
  var SMOKE_RES = 0.5;

  var puffs = [puffSprite(), puffSprite(), puffSprite()];
  var emberSprite = {
    o: softDot(ORANGE, 32, [[0, 1], [0.25, 0.9], [1, 0]]),
    r: softDot(RED, 32, [[0, 1], [0.25, 0.85], [1, 0]])
  };
  var w = 0, h = 0, dpr = 1, raf = 0, last = 0, emberDebt = 0, cloudTimer = 0;
  var clouds = [], embers = [], glow = null;

  // A lumpy puff: radial falloff broken up by fractal noise, so edges
  // read as smoke instead of a clean circle.
  function puffSprite() {
    var size = 128, c = document.createElement('canvas');
    c.width = c.height = size;
    var g = c.getContext('2d');
    var img = g.createImageData(size, size);
    var octaves = [];
    for (var o = 0, cells = 4; o < 4; o++, cells *= 2) {
      var grid = new Float32Array((cells + 1) * (cells + 1));
      for (var i = 0; i < grid.length; i++) grid[i] = Math.random();
      octaves.push({ cells: cells, grid: grid, amp: Math.pow(0.5, o) });
    }
    function smooth(t) { return t * t * (3 - 2 * t); }
    for (var y = 0; y < size; y++) {
      for (var x = 0; x < size; x++) {
        var n = 0;
        for (var k = 0; k < octaves.length; k++) {
          var oc = octaves[k], cs = oc.cells, row = cs + 1;
          var fx = x / size * cs, fy = y / size * cs;
          var x0 = Math.floor(fx), y0 = Math.floor(fy);
          var tx = smooth(fx - x0), ty = smooth(fy - y0);
          var a = oc.grid[y0 * row + x0], b = oc.grid[y0 * row + x0 + 1];
          var c2 = oc.grid[(y0 + 1) * row + x0], d = oc.grid[(y0 + 1) * row + x0 + 1];
          n += oc.amp * ((a + (b - a) * tx) * (1 - ty) + (c2 + (d - c2) * tx) * ty);
        }
        n /= 1.875;
        var dx = x / size * 2 - 1, dy = y / size * 2 - 1;
        var fall = Math.max(0, 1 - Math.sqrt(dx * dx + dy * dy));
        var v = Math.min(1, Math.max(0, fall * 4 * (n + 0.1) - 0.35));
        var p = (y * size + x) * 4;
        img.data[p] = ASH[0];
        img.data[p + 1] = ASH[1];
        img.data[p + 2] = ASH[2];
        img.data[p + 3] = Math.round(v * 255);
      }
    }
    g.putImageData(img, 0, 0);
    return c;
  }

  function softDot(rgb, size, stops) {
    var c = document.createElement('canvas');
    c.width = c.height = size;
    var g = c.getContext('2d');
    var grd = g.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2);
    stops.forEach(function (s) { grd.addColorStop(s[0], 'rgba(' + rgb + ',' + s[1] + ')'); });
    g.fillStyle = grd;
    g.fillRect(0, 0, size, size);
    return c;
  }

  function rand(a, b) { return a + Math.random() * (b - a); }

  // Thick clouds are tight clusters of large puffs; thin wisps are a few
  // faint puffs stretched along the wind.
  function makeCloud(x) {
    var thick = Math.random() < 0.45;
    var unit = Math.max(w, h) * (thick ? rand(0.1, 0.16) : rand(0.08, 0.13));
    var count = thick ? Math.round(rand(10, 15)) : Math.round(rand(4, 6));
    var blobs = [];
    for (var i = 0; i < count; i++) {
      blobs.push({
        dx: rand(-1.1, 1.1) * unit * (thick ? 1 : 2.4),
        dy: rand(-0.5, 0.5) * unit * (thick ? 1 : 0.5),
        r: unit * rand(0.6, 1.1),
        sx: thick ? rand(1, 1.3) : rand(1.8, 2.6),
        rot: rand(0, 6.28),
        spin: rand(-0.04, 0.04),
        sprite: puffs[i % puffs.length]
      });
    }
    var span = unit * (thick ? 2.6 : 4.4);
    return {
      x: x === undefined ? -span : x,
      y: h * rand(0.1, 1.0),
      vx: WIND * rand(0.7, 1.4),
      vy: -rand(2, 8),
      span: span,
      alpha: thick ? 1 : rand(0.3, 0.5),
      age: 0,
      blobs: blobs
    };
  }

  function resize() {
    dpr = Math.min(window.devicePixelRatio || 1, 1.5);
    var first = w === 0;
    w = window.innerWidth;
    h = window.innerHeight;
    canvas.width = Math.round(w * dpr);
    canvas.height = Math.round(h * dpr);
    smoke.width = Math.round(w * SMOKE_RES);
    smoke.height = Math.round(h * SMOKE_RES);

    glow = ctx.createRadialGradient(w * 0.5, h * 1.15, 0, w * 0.5, h * 1.15, Math.max(w, h) * 0.75);
    glow.addColorStop(0, 'rgba(' + ORANGE + ',.14)');
    glow.addColorStop(0.45, 'rgba(' + RED + ',.04)');
    glow.addColorStop(1, 'rgba(' + RED + ',0)');

    // Start with clouds already on screen instead of an empty sky.
    if (first) {
      var n = Math.round(2 + w / 500);
      for (var i = 0; i < n; i++) {
        var c = makeCloud(rand(-0.1, 1) * w);
        c.age = 10;
        clouds.push(c);
      }
    }
  }

  function drawClouds(dt) {
    cloudTimer -= dt;
    if (cloudTimer <= 0) {
      clouds.push(makeCloud());
      cloudTimer = rand(5, 9) * Math.min(1.5, 1200 / Math.max(w, 600));
    }

    sctx.setTransform(1, 0, 0, 1, 0, 0);
    sctx.clearRect(0, 0, smoke.width, smoke.height);
    for (var i = clouds.length - 1; i >= 0; i--) {
      var c = clouds[i];
      c.age += dt;
      c.x += c.vx * dt;
      c.y += c.vy * dt;
      if (c.x - c.span > w) { clouds.splice(i, 1); continue; }
      // Clouds billow: they grow slowly as they travel.
      var grow = 1 + c.age * 0.012;
      var fadeIn = Math.min(1, c.age / 3);
      sctx.globalAlpha = c.alpha * fadeIn;
      for (var j = 0; j < c.blobs.length; j++) {
        var b = c.blobs[j];
        b.rot += b.spin * dt;
        var rx = b.r * grow * b.sx, ry = b.r * grow;
        var cos = Math.cos(b.rot * 0.2), sin = Math.sin(b.rot * 0.2);
        sctx.setTransform(SMOKE_RES * cos, SMOKE_RES * sin, -SMOKE_RES * sin, SMOKE_RES * cos,
          SMOKE_RES * (c.x + b.dx * grow), SMOKE_RES * (c.y + b.dy * grow));
        sctx.drawImage(b.sprite, -rx, -ry, rx * 2, ry * 2);
      }
    }

    ctx.globalAlpha = SMOKE_MAX;
    ctx.drawImage(smoke, 0, 0, w, h);
    ctx.globalAlpha = 1;
  }

  function spawnEmber() {
    // Most enter from the left on the wind, some lift from the bottom.
    var fromLeft = Math.random() < 0.6;
    embers.push({
      x: fromLeft ? -10 : Math.random() * w,
      y: fromLeft ? h * (0.35 + Math.random() * 0.65) : h + 10,
      vx: WIND * (1.2 + Math.random() * 2),
      vy: -(12 + Math.random() * 40),
      r: 1.2 + Math.random() * 1.8,
      age: 0,
      life: 6 + Math.random() * 8,
      phase: Math.random() * 6.28,
      hot: Math.random() < 0.65
    });
  }

  function drawEmbers(dt) {
    emberDebt += dt * (2 + w / 260);
    while (emberDebt >= 1) { spawnEmber(); emberDebt -= 1; }
    for (var i = embers.length - 1; i >= 0; i--) {
      var e = embers[i];
      e.age += dt;
      var k = e.age / e.life;
      if (k >= 1 || e.x > w + 20 || e.y < -20) { embers.splice(i, 1); continue; }
      // Gusts: sideways and vertical wobble on different clocks
      e.x += (e.vx + Math.sin(e.phase + e.age * 1.3) * 14) * dt;
      e.y += (e.vy + Math.cos(e.phase + e.age * 2.1) * 10) * dt;
      var fade = Math.min(1, e.age / 0.6) * (1 - k);
      var flicker = 0.65 + 0.35 * Math.sin(e.phase + e.age * 9);
      ctx.globalAlpha = fade * flicker;
      var s = e.r * 4;
      ctx.drawImage(e.hot ? emberSprite.o : emberSprite.r, e.x - s, e.y - s, s * 2, s * 2);
    }
    ctx.globalAlpha = 1;
  }

  function paint(dt, withEmbers) {
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = glow;
    ctx.fillRect(0, 0, w, h);
    drawClouds(dt);
    if (withEmbers) drawEmbers(dt);
  }

  function frame(now) {
    var dt = Math.min((now - last) / 1000, 0.05);
    last = now;
    paint(dt, true);
    raf = requestAnimationFrame(frame);
  }

  function start() {
    if (raf) return;
    if (reduceMotion.matches) { paint(0, false); return; }
    last = performance.now();
    raf = requestAnimationFrame(frame);
  }

  function stop() {
    cancelAnimationFrame(raf);
    raf = 0;
  }

  window.addEventListener('resize', function () {
    resize();
    if (!raf) paint(0, false);
  });
  document.addEventListener('visibilitychange', function () {
    if (document.hidden) stop(); else start();
  });
  if (reduceMotion.addEventListener) {
    reduceMotion.addEventListener('change', function () { stop(); embers.length = 0; start(); });
  }

  resize();
  start();
})();
