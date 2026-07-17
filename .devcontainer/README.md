# Dev Container

This repository includes a VS Code Dev Container configuration in `.devcontainer/devcontainer.json`.

## Usage

1. Open this folder in VS Code.
2. Run **Dev Containers: Reopen in Container**.
3. Once the container is running, you can preview the site:
   - Use the Live Server extension (default port `5500`), or
   - Run `python3 -m http.server 8000` in the terminal.

## Tailwind Development

- Run `npm run dev` to start both:
  - Tailwind CSS watcher
  - Local web server at `http://localhost:8000`
- Run `npm run build:css` for a production CSS build.

## Notes

- The container uses `mcr.microsoft.com/devcontainers/base:ubuntu`.
- Ports `5500` and `8000` are forwarded for local preview.
