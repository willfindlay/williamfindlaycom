(function () {
  "use strict";

  var COLLAPSED_HEIGHT = 160; // px, matches max-height in CSS (10rem)

  function initCopyButtons() {
    var blocks = document.querySelectorAll(".code-block");
    blocks.forEach(function (block) {
      if (block.querySelector(".code-block__copy")) return;

      var btn = document.createElement("button");
      btn.className = "code-block__copy";
      btn.textContent = "Copy";
      btn.setAttribute("aria-label", "Copy code to clipboard");

      btn.addEventListener("click", function () {
        var pre = block.querySelector("pre");
        if (!pre) return;
        navigator.clipboard.writeText(pre.textContent).then(function () {
          btn.textContent = "Copied!";
          btn.classList.add("code-block__copy--copied");
          setTimeout(function () {
            btn.textContent = "Copy";
            btn.classList.remove("code-block__copy--copied");
          }, 2000);
        });
      });

      block.appendChild(btn);
    });
  }

  function initCollapsible() {
    var blocks = document.querySelectorAll(".code-block[data-collapsed]");
    blocks.forEach(function (block) {
      if (block.querySelector(".code-block__toggle")) return;

      var pre = block.querySelector("pre");
      if (!pre) return;

      // Measure natural height before collapsing. The pre has no max-height
      // yet, so scrollHeight == clientHeight == natural height.
      var naturalHeight = pre.scrollHeight;
      if (naturalHeight <= COLLAPSED_HEIGHT + 20) return;

      block.classList.add("code-block--collapsed");

      var toggle = document.createElement("button");
      toggle.className = "code-block__toggle";
      toggle.textContent = "Show more";
      toggle.setAttribute("aria-expanded", "false");

      toggle.addEventListener("click", function () {
        var isCollapsed = block.classList.toggle("code-block--collapsed");
        toggle.textContent = isCollapsed ? "Show more" : "Show less";
        toggle.setAttribute("aria-expanded", String(!isCollapsed));
      });

      block.appendChild(toggle);
    });
  }

  function init() {
    initCopyButtons();
    initCollapsible();
  }

  init();

  document.addEventListener("spa:navigate", function () {
    init();
  });
})();
