# MyController v2.x Web Console
Web console source for the MyController server. Dev and production builds use [Vite](https://vite.dev/).

This directory is part of the server repository. Production builds are packed into the server binary (`go run ./scripts/cmd/pack_web_console`) and extracted to disk on first start.

## Setup development environment
Run the MyController server from this repository (or another reachable instance).

By default the dev server proxies HTTP and websocket to `localhost:8080`.

*   `MC_PROXY_HTTP` - MyController server http url, default: `http://localhost:8080`
*   `MC_PROXY_WEBSOCKET` - MyController server websocket url, default `ws://localhost:8080`

```bash
# example: MyController server is running on 192.168.1.21
export MC_PROXY_HTTP="http://192.168.1.21:8080"
export MC_PROXY_WEBSOCKET="ws://192.168.1.21:8080"

cd web-console
corepack enable
yarn install
yarn start
```

## Available Scripts
In the project directory, you can run:

### `yarn start`
Runs the app in the development mode.<br />
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.<br />
You will also see any lint errors in the console.

### `yarn build`
Builds the app for production to the `build` folder.<br />
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br />

For a packed server binary, leave `web.web_directory` empty so the UI is extracted on first start. To serve this `build/` folder from disk instead, set `web.web_directory` to it.

## Locales
Edit translation files in `public/locales/<lang>.yaml` and commit them here. The UI loads those files at runtime (`/locales/{{lng}}.yaml`).

## Release
The console is released with the server. CI packs `web-console/build` into the server binary. The process extracts it to `{data}/internal/web_console` on first start so the UI is not kept in RAM.
