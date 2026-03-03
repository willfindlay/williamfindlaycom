(function () {
  "use strict";

  var header = document.querySelector(".site-header");
  var reducedMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)"
  ).matches;
  var ticking = false;

  function getArticle() {
    return document.querySelector(".post");
  }

  function update() {
    ticking = false;
    var article = getArticle();
    if (!article || !header) {
      return;
    }

    var rect = article.getBoundingClientRect();
    var articleTop = rect.top + window.scrollY;
    var scrollable = article.offsetHeight - window.innerHeight;

    if (scrollable <= 0) {
      header.style.setProperty("--reading-progress", "1");
      return;
    }

    var progress = (window.scrollY - articleTop) / scrollable;
    if (progress < 0) progress = 0;
    if (progress > 1) progress = 1;

    header.style.setProperty("--reading-progress", progress);
  }

  function onScroll() {
    if (!ticking) {
      requestAnimationFrame(update);
      ticking = true;
    }
  }

  function bind() {
    if (reducedMotion) return;
    if (getArticle()) {
      window.addEventListener("scroll", onScroll, { passive: true });
      update();
    } else {
      window.removeEventListener("scroll", onScroll);
      if (header) {
        header.style.removeProperty("--reading-progress");
      }
    }
  }

  bind();

  document.addEventListener("spa:navigate", function () {
    bind();
  });
})();
