import path from "path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA, VitePWAOptions } from "vite-plugin-pwa";
import { register } from "module";


const manifestForPlugIn:VitePWAOptions = {
  includeAssets: ['favicon.ico', "apple-touch-icon.png", "masked-icon.svg"],
  manifest: {
    name: "SudoCam",
    short_name: "SudoCam",
    description: "SudoCam is a camera management system",
    // icons: [{
    //   src: '/android-chrome-192x192.png',
    //   sizes: '192x192',
    //   type: 'image/png',
    //   purpose: 'favicon'
    // },
    // {
    //   src: '/android-chrome-512x512.png',
    //   sizes: '512x512',
    //   type: 'image/png',
    //   purpose: 'favicon'
    // },
    // {
    //   src: '/apple-touch-icon.png',
    //   sizes: '180x180',
    //   type: 'image/png',
    //   purpose: 'apple touch icon',
    // },
    // {
    //   src: '/maskable_icon.png',
    //   sizes: '512x512',
    //   type: 'image/png',
    //   purpose: 'any maskable',
    // }
    // ],
    theme_color: '#171717',
    background_color: '#f0e7db',
    display: "standalone",
    scope: '/',
    start_url: "/",
    orientation: 'portrait-primary'
  },
  injectRegister: "inline",
  minify: false,
  workbox: undefined,
  injectManifest: undefined,
  includeManifestIcons: false,
  disable: false
}

export default defineConfig(({ mode, command })=> {
  const { SUDOCAM_PROXY_URL } = process.env;
  console.log("SUDOCAM_PROXY_URL", SUDOCAM_PROXY_URL);
  return {
    plugins: [react()],//
    base: "/",
    appType: "spa",

    server: {
      host: "0.0.0.0",
      proxy: SUDOCAM_PROXY_URL ? {
        '/api': {
          target: SUDOCAM_PROXY_URL, // Your backend URL
          changeOrigin: true,
          // secure: false, // If your backend is not using HTTPS in development
          // cookieDomainRewrite: 'localhost', // Careful with this
        },
        '/api/ws': {
          target: 'ws://localhost:8080', // Your backend WebSocket URL (ws:// or wss://)
          ws: true,
          changeOrigin: true, // Required for proper proxying
        },
      } : undefined
    },
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
  }
});
