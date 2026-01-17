function preferredTheme() {
  if (document.documentElement.classList.contains("dark")) {
    return "dark";
  }
  if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }
  return "default";
}

function initMermaid() {
  const theme = preferredTheme();
  window.mermaid.initialize({ startOnLoad: true, theme });
  window.mermaid.init(undefined, document.querySelectorAll(".mermaid"));
}

function loadMermaid() {
  if (window.mermaid) {
    initMermaid();
    return;
  }

  const script = document.createElement("script");
  script.src = "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js";
  script.onload = () => {
    if (window.mermaid) {
      initMermaid();
    }
  };
  document.head.appendChild(script);
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", loadMermaid);
} else {
  loadMermaid();
}
