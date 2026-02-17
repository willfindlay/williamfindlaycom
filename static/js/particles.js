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

  // Formation
  const FORMATION_TEXT = canvas.dataset.formation || "";
  const PHASE = { IDLE: 0, GATHERING: 1, HOLDING: 2, RELEASING: 3 };
  const GATHER_MS = 2000;
  const HOLD_MS = 1500;
  const RELEASE_MS = 3000;
  let formationPhase = PHASE.IDLE;
  let formationStart = 0;
  let formationCount = 0;

  let width, height, dpr;
  let wrapW, wrapH;
  let mouse = { x: -1000, y: -1000 };
  let particles = [];
  let animId;

  function cancelFormation() {
    if (formationPhase === PHASE.IDLE) return;
    formationPhase = PHASE.IDLE;
    formationCount = 0;
    for (let i = 0; i < particles.length; i++) {
      const p = particles[i];
      p.isFormation = false;
      p.targetX = undefined;
      p.targetY = undefined;
      p.originX = undefined;
      p.originY = undefined;
      p.releaseX = undefined;
      p.releaseY = undefined;
      p.formationNeighbors = undefined;
    }
  }

  function resize() {
    cancelFormation();

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

  // --- Formation logic ---

  function sampleTextPoints(text, fontSize) {
    const off = document.createElement("canvas");
    const offCtx = off.getContext("2d");
    const font = `bold ${fontSize}px "DejaVu Sans", sans-serif`;
    offCtx.font = font;
    const metrics = offCtx.measureText(text);

    const tw = Math.ceil(metrics.width);
    const th = Math.ceil(fontSize * 1.4);
    off.width = tw;
    off.height = th;

    offCtx.font = font;
    offCtx.fillStyle = "#fff";
    offCtx.textBaseline = "middle";
    offCtx.fillText(text, 0, th / 2);

    const data = offCtx.getImageData(0, 0, tw, th).data;
    const points = [];
    const step = Math.max(3, Math.round(fontSize / 40));

    for (let y = 0; y < th; y += step) {
      for (let x = 0; x < tw; x += step) {
        if (data[(y * tw + x) * 4 + 3] > 128) {
          points.push({ x: x - tw / 2, y: y - th / 2 });
        }
      }
    }

    return points;
  }

  function easeOutCubic(t) {
    return 1 - (1 - t) * (1 - t) * (1 - t);
  }

  function initFormation() {
    if (reducedMotion || width < 768 || !FORMATION_TEXT) return;

    const fontSize = Math.min(width * 0.08, 130);
    const points = sampleTextPoints(FORMATION_TEXT, fontSize);

    const targetCount = Math.min(
      Math.floor(particles.length * 0.7),
      85,
      points.length,
    );

    // Stride selection: pick every Nth point for even spatial distribution
    const stride = Math.max(1, Math.floor(points.length / targetCount));
    const selected = [];
    for (let i = 0; i < points.length && selected.length < targetCount; i += stride) {
      selected.push(points[i]);
    }
    if (selected.length < 25) return;

    // Anchor WF relative to the hero title
    let cx, cy;
    const titleEl = document.querySelector(".hero__title");
    if (titleEl) {
      const rect = titleEl.getBoundingClientRect();
      if (width >= 768) {
        // Desktop: to the right of the title
        cx = rect.right + 60;
        cy = rect.top + rect.height / 2;
      } else {
        // Mobile: centered behind the title
        cx = rect.left + rect.width / 2;
        cy = rect.top + rect.height / 2;
      }
    } else {
      cx = width * 0.72;
      cy = height * 0.28;
    }

    formationCount = selected.length;
    for (let i = 0; i < selected.length; i++) {
      const p = particles[i];
      p.originX = p.x;
      p.originY = p.y;
      p.targetX = cx + selected[i].x;
      p.targetY = cy + selected[i].y;
      p.isFormation = true;
    }

    // Precompute 3 nearest neighbors by target position for mesh connections
    for (let i = 0; i < formationCount; i++) {
      const pi = particles[i];
      const neighbors = [];
      for (let j = 0; j < formationCount; j++) {
        if (i === j) continue;
        const dx = pi.targetX - particles[j].targetX;
        const dy = pi.targetY - particles[j].targetY;
        neighbors.push({ idx: j, d: dx * dx + dy * dy });
      }
      neighbors.sort((a, b) => a.d - b.d);
      pi.formationNeighbors = [neighbors[0].idx, neighbors[1].idx, neighbors[2].idx];
    }

    formationPhase = PHASE.GATHERING;
    formationStart = performance.now();
  }

  // --- End formation logic ---

  function update(now) {
    // Formation phase transitions
    if (formationPhase === PHASE.GATHERING) {
      if (now - formationStart >= GATHER_MS) {
        formationPhase = PHASE.HOLDING;
        formationStart = now;
      }
    } else if (formationPhase === PHASE.HOLDING) {
      if (now - formationStart >= HOLD_MS) {
        formationPhase = PHASE.RELEASING;
        formationStart = now;
        for (let i = 0; i < particles.length; i++) {
          const p = particles[i];
          if (p.isFormation) {
            p.releaseX = p.x;
            p.releaseY = p.y;
          }
        }
      }
    } else if (formationPhase === PHASE.RELEASING) {
      if (now - formationStart >= RELEASE_MS) {
        cancelFormation();
      }
    }

    for (let i = 0; i < particles.length; i++) {
      const p = particles[i];

      // Pulse
      p.pulsePhase = (p.pulsePhase + CONFIG.pulseSpeed) % (Math.PI * 2);
      const pulse = Math.sin(p.pulsePhase);
      p.radius = p.baseRadius + pulse * 0.5;
      p.opacity = 0.3 + pulse * 0.15;

      // Formation: gathering
      if (p.isFormation && formationPhase === PHASE.GATHERING) {
        const t = easeOutCubic(
          Math.min((now - formationStart) / GATHER_MS, 1),
        );
        p.x = p.originX + (p.targetX - p.originX) * t;
        p.y = p.originY + (p.targetY - p.originY) * t;
        continue;
      }

      // Formation: holding (Lissajous jitter)
      if (p.isFormation && formationPhase === PHASE.HOLDING) {
        const jitter = 1.5;
        p.x = p.targetX + Math.sin(p.pulsePhase * 1.3) * jitter;
        p.y = p.targetY + Math.cos(p.pulsePhase * 0.9) * jitter;
        continue;
      }

      // Formation: releasing (ease back to origin)
      if (p.isFormation && formationPhase === PHASE.RELEASING) {
        const t = easeOutCubic(
          Math.min((now - formationStart) / RELEASE_MS, 1),
        );
        p.x = p.releaseX + (p.originX - p.releaseX) * t;
        p.y = p.releaseY + (p.originY - p.releaseY) * t;
        continue;
      }

      // Normal physics (IDLE or non-formation particles)
      const dx = p.x - mouse.x;
      const dy = p.y - mouse.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      if (dist < CONFIG.mouseRadius && dist > 0) {
        const force = (1 - dist / CONFIG.mouseRadius) * CONFIG.mouseForce;
        p.vx += (dx / dist) * force;
        p.vy += (dy / dist) * force;
      }

      p.x += p.vx;
      p.y += p.vy;

      p.vx *= 0.998;
      p.vy *= 0.998;

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
      const formationActive =
        formationPhase === PHASE.GATHERING || formationPhase === PHASE.HOLDING;
      const maxDist = CONFIG.connectionDistance;

      // Distance-based connections (skip both-formation pairs during formation)
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const a = particles[i];
          const b = particles[j];

          if (formationActive && a.isFormation && b.isFormation) continue;

          const dx = a.x - b.x;
          const dy = a.y - b.y;
          const dist = Math.sqrt(dx * dx + dy * dy);

          if (dist < maxDist) {
            const opacity =
              (1 - dist / maxDist) * CONFIG.connectionOpacity;
            ctx.beginPath();
            ctx.moveTo(a.x, a.y);
            ctx.lineTo(b.x, b.y);
            ctx.strokeStyle = `rgba(${CONFIG.color.r}, ${CONFIG.color.g}, ${CONFIG.color.b}, ${opacity})`;
            ctx.lineWidth = 0.5;
            ctx.stroke();
          }
        }
      }

      // Nearest-neighbor connections between formation particles (distance-capped)
      if (formationActive && formationCount > 1) {
        const { r, g, b } = CONFIG.color;
        ctx.lineWidth = 0.8;
        for (let i = 0; i < formationCount; i++) {
          const a = particles[i];
          if (!a.formationNeighbors) continue;
          for (const j of a.formationNeighbors) {
            if (j < i) continue; // each line drawn once
            const nb = particles[j];
            const dx = a.x - nb.x;
            const dy = a.y - nb.y;
            const dist = Math.sqrt(dx * dx + dy * dy);
            if (dist < maxDist) {
              const opacity = (1 - dist / maxDist) * 0.25;
              ctx.beginPath();
              ctx.moveTo(a.x, a.y);
              ctx.lineTo(nb.x, nb.y);
              ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, ${opacity})`;
              ctx.stroke();
            }
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

  function loop(now) {
    if (!reducedMotion) {
      update(now);
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
  if (FORMATION_TEXT) {
    document.fonts.ready.then(() => {
      setTimeout(initFormation, 500);
    });
  }
  loop(performance.now());
})();
