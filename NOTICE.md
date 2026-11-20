# Third-party content

This repository's code is under the [MIT License](LICENSE). It contains no image from the galleries below, only the public domain backgrounds listed after them; the app fetches them from the galleries at run time, each only when the user browses or uses it.

## POSE-DB-01 and COSTUME-DB-01

<https://ai-pose-gallery.pages.dev>, <https://ai-costume-gallery.pages.dev>

Their terms, summarized from the sites (the sites' Japanese text is authoritative):

- Free to use, including commercial use, editing and reference in your own works; credit is optional.
- Redistributing, reposting or selling the image data itself, original or edited, is not allowed.
- Downloading the images in bulk, including automated downloading with tools, is not allowed; using them one at a time as references for generation is.
- They must not be used to depict real people or minors, or for anything offensive.
- They are provided without warranty. The outfit gallery notes that its outfits are generated from text and checked for similarity to existing brands, without a guarantee that they infringe no one's rights.

How the app follows these terms: it downloads only the full images the user chose, when the user presses Generate, one each; it keeps no copy beyond the platform's image cache; and it takes no photo of a person, so their images only ever go on a fictional person made from a pose, an outfit or the user's own clothes, and a CC0 face (`Request` in `apps/mirra/generate.go`). Photos the user adds in the wardrobe are of clothes, laid flat or on a hanger, and stay on the device.

## FACE REFERENCE

<https://ai-portrait-gallery.pages.dev>

AI-generated faces of people who do not exist, released under [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/). The site asks that they are not used to impersonate real people or for anything offensive.

## Backgrounds

`apps/mirra/backgrounds/` holds the scenes a look can be set in: photographs with no one in them, each released into the public domain under [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/) and found through [Openverse](https://openverse.org). The app ships them scaled to 960 pixels with their metadata removed, and sends the chosen one with the request.

| File | Source |
| --- | --- |
| `indoor-brick-wall.jpg` | [StockSnap](https://stocksnap.io/photo/bricks-wall-ORBN2T3893) |
| `indoor-white-counter.jpg` | [StockSnap](https://stocksnap.io/photo/cafe-window-NDIQFBYIHR) |
| `indoor-empty-room.jpg` | [rawpixel](https://www.rawpixel.com/image/8732262/photo-image-public-domain-wood-house) |
| `indoor-sunny-room.jpg` | [rawpixel](https://www.rawpixel.com/image/5903577/photo-image-public-domain-house-living-room) |
| `indoor-bright-loft.jpg` | [rawpixel](https://www.rawpixel.com/image/3283151/free-photo-image-room-living-space-coworking) |
| `city-tree-street.jpg` | [StockSnap](https://stocksnap.io/photo/city-street-LAVXQK6AUP) |
| `city-skyline-day.jpg` | [StockSnap](https://stocksnap.io/photo/cityscape-skyline-7YHOOBJPHE) |
| `city-skyline-dusk.jpg` | [StockSnap](https://stocksnap.io/photo/city-skyline-HVKT6YLKOJ) |
| `city-skyline-night.jpg` | [StockSnap](https://stocksnap.io/photo/buildings-skyline-EFMLJD0R2I) |
| `nature-autumn-path.jpg` | [rawpixel](https://www.rawpixel.com/image/6040003/photo-image-background-aesthetic-public-domain) |
| `nature-lavender.jpg` | [StockSnap](https://stocksnap.io/photo/flower-field-JVBWZNCQLW) |
| `nature-forest.jpg` | [rawpixel](https://www.rawpixel.com/image/8809263/photo-image-plant-light-tree) |
| `nature-snow-mountain.jpg` | [StockSnap](https://stocksnap.io/photo/snow-mountain-0KHG3WOP1F) |
| `nature-rocky-beach.jpg` | [StockSnap](https://stocksnap.io/photo/beach-ocean-VV40UIW8KF) |

## Your responsibility

What you generate with this app is yours to answer for: the galleries' terms above, the terms of the image API you configure, and the law where you are. Do not use it to depict or impersonate real people without their consent, or to depict minors.
