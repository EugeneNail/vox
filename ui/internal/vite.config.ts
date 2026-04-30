import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: ["vox-app.site"],
    host: "0.0.0.0",
    hmr: false,
    watch: null,
    port: 5173,
  },
});
