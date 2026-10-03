/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./ui/html/**/*.tmpl", "./ui/static/js/**/*.js"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "surface-container-high": "#dce0e7",
        "on-primary-container": "#bec8ff",
        outline: "#747686",
        "on-surface": "#131b2e",
        surface: "#e6e8ec",
        "surface-container": "#e4e7ec",
        "outline-variant": "#b8bcc6",
        primary: "#1A45D2",
        accent: "#B33A18",
        "surface-container-low": "#eceef2",
        "primary-container": "#1a45d2",
        "accent-container": "#B33A18",
        error: "#ba1a1a",
        "surface-container-lowest": "#e6e8ec",
        "on-surface-variant": "#444655",
        "on-background": "#131b2e",
        "primary-fixed-dim": "#b8c3ff",
        "node-green": "#2e7d32",
        "node-purple": "#7b1fa2",
      },
      borderRadius: {
        DEFAULT: "0.125rem",
        lg: "0.25rem",
        xl: "0.5rem",
        full: "9999px",
      },
      spacing: {
        "margin-desktop": "64px",
        gutter: "24px",
      },
      fontFamily: {
        headline: ["Geist"],
        label: ["JetBrains Mono"],
        body: ["Geist"],
      },
      fontSize: {
        "headline-md": ["24px", { lineHeight: "32px", fontWeight: "600" }],
        "display-lg": [
          "48px",
          { lineHeight: "56px", letterSpacing: "-0.02em", fontWeight: "700" },
        ],
        "display-xl": ["180px", { lineHeight: "1", fontWeight: "500" }],
        "headline-lg": ["32px", { lineHeight: "40px", fontWeight: "600" }],
        "label-md": [
          "14px",
          { lineHeight: "20px", letterSpacing: "0.01em", fontWeight: "500" },
        ],
        "body-lg": ["18px", { lineHeight: "28px", fontWeight: "400" }],
        "label-sm": [
          "12px",
          { lineHeight: "16px", letterSpacing: "0.05em", fontWeight: "600" },
        ],
        "body-md": ["16px", { lineHeight: "24px", fontWeight: "400" }],
        "body-sm": ["14px", { lineHeight: "20px", fontWeight: "400" }],
      },
    },
  },
  plugins: [require("@tailwindcss/forms")],
};
