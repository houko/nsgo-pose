# Mirra architecture

Mirra is an NSGo app for macOS, iOS, Android and Windows that puts a face, an outfit and a pose from three public reference galleries together and has an OpenAI-compatible image model, in the cloud or on the user's own server, draw the person they describe. It is a single Go module in `apps/mirra`; its website is in `apps/website`.

## Code map

| Path | Responsibility |
| --- | --- |
| `apps/mirra/go.mod`, `apps/mirra/nsgo.yaml` | The only Go module and the packaging configuration; run Go and the `nsgo` tool in `apps/mirra`. |
| `apps/mirra/app.go` | Entry point; the app's state, the choice of layout by window width, generating, saving, starring and deleting a picture. |
| `apps/mirra/gallery.go` | Reading the three galleries' indexes live, what is left out, and the chips each filters by. |
| `apps/mirra/generate.go` | The request: the prompt, the reference images in order and the `/images/edits` call. |
| `apps/mirra/service.go` | Cloud or local service and the connection test against `/models`. |
| `apps/mirra/works.go` | The lookbook: pictures named by when they were made, with a JSON sidecar for their number, star and look. |
| `apps/mirra/wardrobe.go` | Favorite outfits, photos of the user's own clothes, and what the look is wearing. |
| `apps/mirra/library.go` | What both layouts share: labels, a gallery's chips and grid, the look's rows, the service's status and Generate. |
| `apps/mirra/phone.go` | The phone layout: Studio, its three-step editor and the tab bar. |
| `apps/mirra/desktop.go` | The desktop studio: top bar, library, zoomable canvas or result, and the look's summary. |
| `apps/mirra/pages.go` | A result, Works and the AI service settings. |
| `apps/mirra/theme.yaml` | The dark amber theme the app always shows. |
| `apps/mirra/locales/` | NSGo message catalogs in English and Simplified Chinese, and the accessors the `nsgo` tool writes from them. |
| `apps/mirra/assets/`, `apps/mirra/tools/appicon` | The app icon and the program that draws it. |
| `apps/mirra/*_test.go` | Headless and target tests; `fixtures_test.go` fakes the galleries and the image API. |
| `screenshots/` | The README's and the website's screenshots. |
| `apps/website/` | The website: TanStack Start prerendered to static pages in Chinese, English and Japanese, published by `.github/workflows/website.yml`. |
| `apps/website/src/i18n/copy.ts` | The website's words in each language. |
| `apps/website/src/components/` | The website's sections, from the hero to the FAQ. |
| `apps/website/vite.config.ts` | The build: the site's address and Pages path, the pages it prerenders and the sitemap, and the version, stars, contributors and latest release it reads. |
| `apps/website/src/i18n/head.ts` | Each page's title, description, canonical and alternate links, Open Graph and Twitter cards and structured data. |
| `apps/website/src/routes/og.$lang.tsx`, `apps/website/scripts/og.sh` | The 1200x630 card each language's Open Graph image is taken from, served by the dev server only, and the script that takes them into `public/`. |

## Boundaries

- Ship no gallery image. Read each gallery's index live, show thumbnails from the gallery while browsing, and download a full image only for the request that uses it.
- Never put a person's photo in a request: the pose and outfit galleries' terms forbid depicting real people with their images. Faces come from the CC0 face gallery; photos the user adds are of clothes.
- Leave out the pose gallery's two-person and explicit lines and the outfit gallery's lingerie and "sexy" categories (`gallery.go`).
- Do network and file work on a goroutine and hand results back with `cx.Post`; never block a builder or a handler.
- Keep API keys in `cx.Secrets()`, never in storage.
- Take colors from the theme. Choose the layout from the window's width (`wideWidth`), not the device.
- Keep the website to what the app does, and show on it no pose or outfit gallery image: its pictures are CC0 faces and the app's own screenshots.

## Core flows

**A look.** The galleries load as resources at start. Choosing a face, an outfit and a pose sets the studio's three signals; only the pose is required.

**Generate.** Without a configured service, Generate opens the settings. Otherwise it downloads the chosen full images, sends them with the prompt to `/images/edits` and writes the picture to the results folder, where it waits. Generate again and the outfit swap make the same request again, the swap with another outfit. Cancel stops the request through its context.

**Save.** Saving moves the picture into the works folder under the next number and writes its sidecar; starring an unsaved picture saves it first. Opening a work brings back its look.

**Service.** The service is a cloud API, which needs a key, or a local server, which may not. The connection test asks `/models`; changing the address, key or kind marks the service untested again.

**Wardrobe.** Favorites are kept in storage as JSON. Added clothes are copied into the uploads folder and numbered by age; their full image is read from disk when generating.
