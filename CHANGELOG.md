# Changelog

## Unreleased

- The AI settings show the provider and key first, with the address, model and size under Advanced and a reset to the provider's defaults; the size is chosen as portrait, square or landscape, the key is shown with an eye button, the About section gives the version, and the desktop settings scroll when they are taller than the window.
- The AI settings preset OpenAI, Google Gemini, Volcengine Ark, xAI, OpenRouter, Alibaba Cloud Model Studio, Azure OpenAI and stable-diffusion.cpp, with their logos and a link to get a key, besides any OpenAI-compatible address. Each provider keeps its own key and settings; the cloud or local service set up before carries over.

## 0.5.0 - 2026-11-20

- Lingerie and sexy outfits, and revealing and two-person poses, show only behind each gallery's 18+ switch, which first asks whether you are 18 or older; the lingerie and sexy chips are translated.

## 0.4.0 - 2026-11-19

- Releases on GitHub with a signed and notarized macOS disk image, a signed Windows installer and the web build, made from `main` when its commits call for a version; the website's download cards link them.
- The app is named Mirra, capitalized, on the web page and in the installed apps.
- A web version at https://yldm-ai.github.io/mirra/app/, published with the website; its key, works and clothes last as long as the page.
- A website at https://yldm-ai.github.io/mirra/ in Chinese, English and Japanese, with share cards, language alternates and a sitemap for search engines.
- The Chinese README's phone screenshot, which was blank, shows the studio again.

## 0.3.0

- A dark amber look with serif titles, after the second reference design.
- The AI service is a cloud API or a local server (whose key may be left out), with a connection test and its status at the top of the app; on the desktop it is set up in a dialog.
- A new picture waits to be saved to Works; generate it again, swap its outfit with one tap, star, export or delete it.
- Works are numbered, can be starred and filtered, and open again with the look they were made from.
- The wardrobe keeps favorite outfits and photos of the user's own clothes, and shows what the look is wearing.
- The desktop canvas zooms, and changes the face or pose from its corner.
- The picture count is gone from the interface; one Generate makes one picture.

## 0.2.0

- Renamed Mirra (镜衣), with a warm theme in light and dark.
- A phone layout (Studio, Wardrobe, Works, Me; the look put together in three steps) and a desktop studio of three columns, chosen by the window's width.
- Faces filter by hair length and come from the face gallery's front-facing section only; poses filter by camera angle.
- Ask for 1, 2 or 4 pictures at once, and cancel a request while it runs.
- Photo mode is gone: the app takes no photo of the user, as the pose and outfit galleries' terms forbid depicting real people with their images.

## 0.1.0

- Combine a pose, an outfit and a face from three reference galleries into one picture through an OpenAI-compatible `/images/edits` endpoint.
- Give your own photo a face from the face gallery; outfits go on poses only, as the outfit gallery's terms require.
- Works kept on the device, with save, share and delete; the API key in the system keychain.
- English and Simplified Chinese.
- macOS, iOS, Android and Windows builds.
