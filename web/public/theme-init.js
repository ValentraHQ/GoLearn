// Applies the saved theme before first paint to avoid a flash. External file so CSP can stay script-src 'self'.
try {
  var t = localStorage.getItem("golearn.theme");
  var dark = t === "dark" || ((!t || t === "system") && matchMedia("(prefers-color-scheme: dark)").matches);
  document.documentElement.classList.toggle("dark", dark);
} catch (e) {}
