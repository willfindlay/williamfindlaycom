(() => {
  function initGiscus() {
    const container = document.querySelector(".giscus-container");
    if (!container) return;

    // Clean up any previous instance (SPA navigation)
    const oldFrame = container.querySelector("iframe.giscus-frame");
    if (oldFrame) oldFrame.remove();
    const oldScript = document.querySelector("script[src*='giscus.app']");
    if (oldScript) oldScript.remove();

    const script = document.createElement("script");
    script.src = "https://giscus.app/client.js";
    script.setAttribute("data-repo", container.dataset.repo);
    script.setAttribute("data-repo-id", container.dataset.repoId);
    script.setAttribute("data-category", container.dataset.category);
    script.setAttribute("data-category-id", container.dataset.categoryId);
    script.setAttribute("data-mapping", "pathname");
    script.setAttribute("data-strict", "0");
    script.setAttribute("data-reactions-enabled", "1");
    script.setAttribute("data-emit-metadata", "0");
    script.setAttribute("data-input-position", "top");
    script.setAttribute("data-theme", container.dataset.theme);
    script.setAttribute("data-lang", "en");
    script.setAttribute("data-loading", "lazy");
    script.crossOrigin = "anonymous";
    script.async = true;

    container.appendChild(script);
  }

  // Initial page load
  initGiscus();

  // SPA navigation
  document.addEventListener("spa:navigate", initGiscus);
})();
