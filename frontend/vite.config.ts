import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [react(), tailwindcss(), wails("./bindings")],
  test: {
    environment: "jsdom",
    // Node 22+'s built-in (experimental) localStorage global shadows
    // jsdom's own implementation with a non-functional stub; disable it so
    // tests get jsdom's working localStorage instead.
    execArgv: ["--no-experimental-webstorage"],
  },
});
