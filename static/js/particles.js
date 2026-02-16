(() => {
  const canvas = document.getElementById("particles");
  if (!canvas) return;

  const ctx = canvas.getContext("2d");
  const reducedMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)",
  ).matches;

  function attr(name, fallback) {
    const v = canvas.dataset[name];
    if (v === undefined) return fallback;
    const n = Number(v);
    return Number.isNaN(n) ? fallback : n;
  }

  function attrColor(name, fallback) {
    const v = canvas.dataset[name];
    if (!v) return fallback;
    const parts = v.split(",").map(Number);
    if (parts.length !== 3 || parts.some(Number.isNaN)) return fallback;
    return { r: parts[0], g: parts[1], b: parts[2] };
  }

  const CONFIG = {
    count: attr("count", 120),
    maxSpeed: attr("speed", 0.3),
    minRadius: attr("sizeMin", 1),
    maxRadius: attr("sizeMax", 2.5),
    connectionDistance: attr("connectDistance", 140),
    connectionOpacity: attr("connectOpacity", 0.08),
    mouseRadius: attr("pushRange", 180),
    mouseForce: attr("pushForce", 0.015),
    pulseSpeed: attr("pulseSpeed", 0.008),
    color: attrColor("color", { r: 79, g: 209, b: 197 }),
    colorAlt: attrColor("colorAlt", { r: 128, g: 90, b: 213 }),
  };

  let width, height, dpr;
  let wrapW, wrapH;
  let mouse = { x: -1000, y: -1000 };
  let particles = [];
  let animId;

  function resize() {
    dpr = Math.min(window.devicePixelRatio || 1, 2);
    width = window.innerWidth;
    height = window.innerHeight;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    canvas.style.width = width + "px";
    canvas.style.height = height + "px";
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    // Expand wrap bounds and spawn particles when viewport grows past max
    if (wrapW && wrapH && (width > wrapW || height > wrapH)) {
      const oldArea = wrapW * wrapH;
      wrapW = Math.max(wrapW, width);
      wrapH = Math.max(wrapH, height);
      const maxParticles = CONFIG.count * 4;
      const extra = Math.round(
        particles.length * ((wrapW * wrapH) / oldArea - 1),
      );
      const toSpawn = Math.min(
        extra,
        Math.max(0, maxParticles - particles.length),
      );
      for (let i = 0; i < toSpawn; i++) {
        const p = createParticle();
        p.baseRadius = p.radius;
        particles.push(p);
      }
    }
  }

  function createParticle() {
    const useAlt = Math.random() < 0.15;
    const c = useAlt ? CONFIG.colorAlt : CONFIG.color;
    return {
      x: Math.random() * (wrapW || width),
      y: Math.random() * (wrapH || height),
      vx: (Math.random() - 0.5) * CONFIG.maxSpeed,
      vy: (Math.random() - 0.5) * CONFIG.maxSpeed,
      radius:
        CONFIG.minRadius +
        Math.random() * (CONFIG.maxRadius - CONFIG.minRadius),
      baseRadius: 0,
      color: c,
      opacity: 0.2 + Math.random() * 0.5,
      pulsePhase: Math.random() * Math.PI * 2,
    };
  }

  function init() {
    wrapW = window.innerWidth;
    wrapH = window.innerHeight;
    resize();
    particles = [];
    for (let i = 0; i < CONFIG.count; i++) {
      const p = createParticle();
      p.baseRadius = p.radius;
      particles.push(p);
    }
  }

  function update() {
    for (let i = 0; i < particles.length; i++) {
      const p = particles[i];

      // Pulse
      p.pulsePhase += CONFIG.pulseSpeed;
      const pulse = Math.sin(p.pulsePhase);
      p.radius = p.baseRadius + pulse * 0.5;
      p.opacity = 0.3 + pulse * 0.15;

      // Mouse repulsion
      const dx = p.x - mouse.x;
      const dy = p.y - mouse.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      if (dist < CONFIG.mouseRadius && dist > 0) {
        const force = (1 - dist / CONFIG.mouseRadius) * CONFIG.mouseForce;
        p.vx += (dx / dist) * force;
        p.vy += (dy / dist) * force;
      }

      // Drift
      p.x += p.vx;
      p.y += p.vy;

      // Damping
      p.vx *= 0.998;
      p.vy *= 0.998;

      // Wrap around edges (fixed bounds so zoom doesn't cluster)
      if (p.x < -10) p.x = wrapW + 10;
      if (p.x > wrapW + 10) p.x = -10;
      if (p.y < -10) p.y = wrapH + 10;
      if (p.y > wrapH + 10) p.y = -10;
    }
  }

  function draw() {
    ctx.clearRect(0, 0, width, height);

    // Connection lines
    if (!reducedMotion) {
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const a = particles[i];
          const b = particles[j];
          const dx = a.x - b.x;
          const dy = a.y - b.y;
          const dist = Math.sqrt(dx * dx + dy * dy);

          if (dist < CONFIG.connectionDistance) {
            const opacity =
              (1 - dist / CONFIG.connectionDistance) * CONFIG.connectionOpacity;
            ctx.beginPath();
            ctx.moveTo(a.x, a.y);
            ctx.lineTo(b.x, b.y);
            ctx.strokeStyle = `rgba(${CONFIG.color.r}, ${CONFIG.color.g}, ${CONFIG.color.b}, ${opacity})`;
            ctx.lineWidth = 0.5;
            ctx.stroke();
          }
        }
      }
    }

    // Particles
    for (let i = 0; i < particles.length; i++) {
      const p = particles[i];
      ctx.beginPath();
      ctx.arc(p.x, p.y, Math.max(p.radius, 0.5), 0, Math.PI * 2);
      ctx.fillStyle = `rgba(${p.color.r}, ${p.color.g}, ${p.color.b}, ${p.opacity})`;
      ctx.fill();
    }
  }

  function loop() {
    if (!reducedMotion) {
      update();
    }
    draw();
    animId = requestAnimationFrame(loop);
  }

  // Event listeners
  window.addEventListener("resize", () => {
    resize();
  });

  window.addEventListener("mousemove", (e) => {
    mouse.x = e.clientX;
    mouse.y = e.clientY;
  });

  window.addEventListener("mouseleave", () => {
    mouse.x = -1000;
    mouse.y = -1000;
  });

  window.addEventListener(
    "touchstart",
    (e) => {
      const t = e.touches[0];
      mouse.x = t.clientX;
      mouse.y = t.clientY;
    },
    { passive: true },
  );

  window.addEventListener(
    "touchmove",
    (e) => {
      const t = e.touches[0];
      mouse.x = t.clientX;
      mouse.y = t.clientY;
    },
    { passive: true },
  );

  window.addEventListener("touchend", () => {
    mouse.x = -1000;
    mouse.y = -1000;
  });

  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      cancelAnimationFrame(animId);
    } else {
      animId = requestAnimationFrame(loop);
    }
  });

  // Start
  init();
  loop();
})();
