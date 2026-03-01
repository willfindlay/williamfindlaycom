(function () {
  "use strict";

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

      // Skip blocks that are already short enough
      if (pre.scrollHeight <= pre.clientHeight + 20) return;

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
