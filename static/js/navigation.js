(() => {
  const SKIP_EXTENSIONS = /\.(xml|pdf|svg|png|jpg|jpeg|gif|zip|tar|gz|json)$/i;
  const TRANSITION_MS = 150;
  const HEAD_SELECTORS = [
    'meta[name="description"]',
    'link[rel="canonical"]',
    'meta[property="og:title"]',
    'meta[property="og:description"]',
    'meta[property="og:type"]',
    'meta[property="og:locale"]',
    'meta[property="og:url"]',
    'meta[property="og:image"]',
    'meta[name="twitter:card"]',
    'meta[name="twitter:title"]',
    'meta[name="twitter:description"]',
    'meta[name="twitter:image"]',
    'script[type="application/ld+json"]',
  ];
  const reducedMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)",
  ).matches;

  let abortController = null;

  history.scrollRestoration = "manual";

  function shouldIntercept(anchor) {
    if (!anchor || !anchor.href) return false;
    if (anchor.target === "_blank") return false;
    if (anchor.hasAttribute("download")) return false;

    const url = new URL(anchor.href, location.origin);

    // External link
    if (url.origin !== location.origin) return false;

    // Non-HTML extension
    if (SKIP_EXTENSIONS.test(url.pathname)) return false;

    // Same page hash link
    if (url.pathname === location.pathname && url.hash) return false;

    return true;
  }

  function syncHead(doc) {
    for (const selector of HEAD_SELECTORS) {
      const oldEl = document.head.querySelector(selector);
      const newEl = doc.head.querySelector(selector);
      if (oldEl) oldEl.remove();
      if (newEl) document.head.appendChild(newEl.cloneNode(true));
    }
  }

  function updateActiveNav(doc) {
    const links = document.querySelectorAll(".nav__links a");
    for (const link of links) {
      const href = link.getAttribute("href");
      const match = doc.querySelector(`.nav__links a[href="${href}"]`);
      if (match && match.classList.contains("active")) {
        link.classList.add("active");
      } else {
        link.classList.remove("active");
      }
    }
  }

  function fadeOut(main) {
    main.style.opacity = "0";
    return new Promise((resolve) => setTimeout(resolve, TRANSITION_MS));
  }

  function fadeIn(main) {
    main.style.opacity = "1";
  }

  async function navigate(url, pushState) {
    if (abortController) abortController.abort();
    abortController = new AbortController();

    const main = document.querySelector("main.main");
    if (!main) {
      location.href = url;
      return;
    }

    // Start fade-out and fetch in parallel
    const fadePromise = reducedMotion ? Promise.resolve() : fadeOut(main);

    let response;
    try {
      response = await fetch(url, {
        signal: abortController.signal,
        headers: { Accept: "text/html" },
      });
    } catch (err) {
      if (err.name === "AbortError") return;
      fadeIn(main);
      location.href = url;
      return;
    }

    if (
      !response.ok ||
      !response.headers.get("content-type")?.includes("text/html")
    ) {
      fadeIn(main);
      location.href = url;
      return;
    }

    let html;
    try {
      html = await response.text();
    } catch {
      fadeIn(main);
      location.href = url;
      return;
    }

    const doc = new DOMParser().parseFromString(html, "text/html");
    const newMain = doc.querySelector("main.main");
    if (!newMain) {
      fadeIn(main);
      location.href = url;
      return;
    }

    // Wait for fade-out to finish before swapping
    await fadePromise;

    // Swap content — safe: HTML is from same-origin server, parsed via DOMParser
    // (DOMParser does not execute scripts or load external resources)
    main.innerHTML = newMain.innerHTML; // eslint-disable-line no-unsanitized/property

    // Update title and head metadata
    document.title = doc.title;
    syncHead(doc);

    // Update active nav
    updateActiveNav(doc);

    // Push state if this is a user click (not popstate)
    if (pushState) {
      history.pushState(null, "", url);
    }

    // Scroll to hash target or top
    const parsed = new URL(url, location.origin);
    if (parsed.hash) {
      const target = document.querySelector(parsed.hash);
      if (target) {
        target.scrollIntoView({ behavior: "instant" });
      } else {
        window.scrollTo({ top: 0, behavior: "instant" });
      }
    } else {
      window.scrollTo({ top: 0, behavior: "instant" });
    }

    // Fade in
    if (!reducedMotion) fadeIn(main);

    document.dispatchEvent(new CustomEvent("spa:navigate"));

    abortController = null;
  }

  // Intercept clicks via event delegation
  document.addEventListener("click", (e) => {
    // Skip modified clicks (new tab / new window)
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    if (e.button !== 0) return;

    const anchor = e.target.closest("a");
    if (!shouldIntercept(anchor)) return;

    e.preventDefault();

    // Same URL — no-op
    if (anchor.href === location.href) return;

    navigate(anchor.href, true);
  });

  // Handle browser back/forward
  window.addEventListener("popstate", () => {
    navigate(location.href, false);
  });
})();
