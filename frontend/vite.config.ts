import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/interfaces": { target: "http://127.0.0.1:8080", changeOrigin: true },
      "/filter": { target: "http://127.0.0.1:8080", changeOrigin: true },
      "/capture": { target: "http://127.0.0.1:8080", changeOrigin: true },
      "/ws": { target: "ws://127.0.0.1:8080", ws: true, changeOrigin: true },
    },
  },
});
