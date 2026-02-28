(function () {
  "use strict";

  var reducedMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)"
  ).matches;

  function reveal(entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add("revealed");
        observer.unobserve(entry.target);
      }
    });
  }

  var observer = new IntersectionObserver(reveal, {
    threshold: 0.1,
    rootMargin: "0px 0px -40px 0px",
  });

  function observe() {
    var elements = document.querySelectorAll("[data-reveal]");
    elements.forEach(function (el, i) {
      if (reducedMotion) {
        el.classList.add("revealed");
        return;
      }
      // Stagger delay: 50ms per item, capped at 300ms
      var delay = Math.min(i * 50, 300);
      el.style.transitionDelay = delay + "ms";
      observer.observe(el);
    });
  }

  // Initial observation
  observe();

  // Re-observe after SPA navigation
  document.addEventListener("spa:navigate", function () {
    observe();
  });
})();
