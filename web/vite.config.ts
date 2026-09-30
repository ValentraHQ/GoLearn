import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: { alias: { "@": new URL("./src", import.meta.url).pathname } },
  server: {
    port: 5173,
    proxy: { "/api": "http://localhost:8080" },
  },
  build: {
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes("@codemirror") || id.includes("@lezer")) return "editor";
          if (id.includes("highlight.js") || id.includes("lowlight")) return "highlight";
          if (id.includes("react-markdown") || id.includes("remark") || id.includes("mdast") || id.includes("micromark") || id.includes("unified") || id.includes("hast")) return "markdown";
          return undefined;
        },
      },
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    globals: true,
    css: false,
  },
});
