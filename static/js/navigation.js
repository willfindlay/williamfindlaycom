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
  let activeTags = new Set();

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

    // Tag toggle (client-side, no server round-trip)
    if (anchor && anchor.dataset.tag && getBlogPosts()) {
      e.preventDefault();
      toggleTag(anchor.dataset.tag);
      return;
    }

    if (!shouldIntercept(anchor)) return;

    e.preventDefault();

    // Same URL, skip navigation
    if (anchor.href === location.href) return;

    navigate(anchor.href, true);
  });

  // Handle browser back/forward
  window.addEventListener("popstate", () => {
    navigate(location.href, false);
  });

  // Client-side blog search

  function escapeHTML(s) {
    return s
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function renderPostCard(post, featured) {
    const tagsHTML = post.tags?.length
      ? `<div class="card__tags">${post.tags.map((t) => `<span class="tag">${escapeHTML(t)}</span>`).join("")}</div>`
      : "";
    return `<article class="card${featured ? " card--featured" : ""}" data-slug="${escapeHTML(post.slug)}">
        <a href="/blog/${escapeHTML(post.slug)}" class="card__link">
            <time class="card__date" datetime="${escapeHTML(post.dateShort)}">${escapeHTML(post.date)}</time>
            <h3 class="card__title">${escapeHTML(post.title)}</h3>
            <p class="card__description">${escapeHTML(post.description)}</p>
            ${tagsHTML}
        </a>
    </article>`;
  }

  function getBlogPosts() {
    const el = document.getElementById("blog-posts-data");
    if (!el) return null;
    try {
      return JSON.parse(el.textContent);
    } catch {
      return null;
    }
  }

  function filterBlogPosts(posts, query) {
    const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
    if (terms.length === 0) return posts;
    return posts.filter((p) => {
      const text = [p.title, p.description, ...(p.tags || [])]
        .join(" ")
        .toLowerCase();
      return terms.every((t) => text.includes(t));
    });
  }

  function initTagsFromURL() {
    activeTags.clear();
    const params = new URLSearchParams(location.search);
    for (const tag of params.getAll("tag")) {
      activeTags.add(tag);
    }
    for (const pill of document.querySelectorAll("[data-tag]")) {
      pill.classList.toggle("tag--active", activeTags.has(pill.dataset.tag));
    }
  }

  function toggleTag(tag) {
    if (activeTags.has(tag)) {
      activeTags.delete(tag);
    } else {
      activeTags.add(tag);
    }
    applyFilters();
  }

  function applyFilters() {
    const posts = getBlogPosts();
    if (!posts) return;

    const postGrid = document.querySelector(".post-grid");
    const emptyState = document.getElementById("empty-state");
    const searchInput = document.querySelector('.search-bar input[name="q"]');

    if (!postGrid) return;

    const query = searchInput ? searchInput.value.trim() : "";

    // Apply text search then tag filter (AND logic)
    let filtered = query ? filterBlogPosts(posts, query) : posts;

    if (activeTags.size > 0) {
      filtered = filtered.filter((p) => {
        const postTags = new Set(p.tags || []);
        for (const t of activeTags) {
          if (!postTags.has(t)) return false;
        }
        return true;
      });
    }

    for (const pill of document.querySelectorAll("[data-tag]")) {
      pill.classList.toggle("tag--active", activeTags.has(pill.dataset.tag));
    }

    const noFilter = activeTags.size === 0 && !query;
    // Values from server-rendered JSON, escaped via escapeHTML
    postGrid.innerHTML = filtered // eslint-disable-line no-unsanitized/property
      .map((p, i) => renderPostCard(p, i === 0 && noFilter))
      .join("");

    if (emptyState) {
      if (filtered.length === 0) {
        if (query && activeTags.size > 0) {
          emptyState.textContent = `No posts matching "${query}" with selected tags.`;
        } else if (query) {
          emptyState.textContent = `No posts matching "${query}".`;
        } else if (activeTags.size > 0) {
          emptyState.textContent = "No posts matching selected tags.";
        } else {
          emptyState.textContent = "No posts yet.";
        }
        emptyState.hidden = false;
      } else {
        emptyState.hidden = true;
      }
    }

    const url = new URL("/blog", location.origin);
    if (query) url.searchParams.set("q", query);
    for (const t of [...activeTags].sort()) {
      url.searchParams.append("tag", t);
    }
    history.replaceState(null, "", url.toString());
  }

  // Delegated input handler for blog search
  document.addEventListener("input", (e) => {
    if (
      e.target.matches('.search-bar input[name="q"]') &&
      document.getElementById("blog-posts-data")
    ) {
      applyFilters();
    }
  });

  // Prevent form submit from doing a full page reload when JS is active
  document.addEventListener("submit", (e) => {
    if (
      e.target.matches(".search-bar") &&
      document.getElementById("blog-posts-data")
    ) {
      e.preventDefault();
      applyFilters();
    }
  });

  // Sync tag state from URL on load and SPA navigation
  initTagsFromURL();
  document.addEventListener("spa:navigate", initTagsFromURL);
})();
