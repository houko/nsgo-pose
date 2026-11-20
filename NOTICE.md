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
| `indoor-gothic-cloister-corridor.jpg` | [StockSnap](https://stocksnap.io/photo/architecture-building-8XS6FA7IQA) |
| `indoor-grand-atrium-library.jpg` | [rawpixel](https://www.rawpixel.com/image/3094430/free-photo-image-school-buildings-america-building) |
| `indoor-modern-open-kitchen.jpg` | [rawpixel](https://www.rawpixel.com/image/5924444/photo-image-light-public-domain-shadow) |
| `indoor-white-loft-lounge.jpg` | [StockSnap](https://stocksnap.io/photo/interior-design-478BZDEA2Y) |
| `indoor-colonial-armory-hall.jpg` | [rawpixel](https://www.rawpixel.com/image/2371395/photo-image-pattern-living-room-public-domain) |
| `indoor-traditional-bedroom-rug.jpg` | [rawpixel](https://www.rawpixel.com/image/6043065/photo-image-public-domain-wood-house) |
| `indoor-palace-grand-staircase.jpg` | [StockSnap](https://stocksnap.io/photo/staircase-interior-HRFCE1Y4MY) |
| `indoor-white-conference-room.jpg` | [StockSnap](https://stocksnap.io/photo/interior-design-70ZAX0HRQQ) |
| `indoor-wooden-hotel-bedroom.jpg` | [StockSnap](https://stocksnap.io/photo/modern-room-ORRHZ1VULX) |
| `indoor-bright-white-kitchen.jpg` | [rawpixel](https://www.rawpixel.com/image/5924974/photo-image-light-public-domain-shadow) |
| `indoor-skyline-home-office.jpg` | [rawpixel](https://www.rawpixel.com/image/5916279/photo-image-background-design-public-domain) |
| `indoor-modern-boardroom-windows.jpg` | [rawpixel](https://www.rawpixel.com/image/3283959/free-photo-image-lobby-conference-room-dining-tabletop) |
| `city-tree-street.jpg` | [StockSnap](https://stocksnap.io/photo/city-street-LAVXQK6AUP) |
| `city-skyline-day.jpg` | [StockSnap](https://stocksnap.io/photo/cityscape-skyline-7YHOOBJPHE) |
| `city-skyline-dusk.jpg` | [StockSnap](https://stocksnap.io/photo/city-skyline-HVKT6YLKOJ) |
| `city-skyline-night.jpg` | [StockSnap](https://stocksnap.io/photo/buildings-skyline-EFMLJD0R2I) |
| `city-burano-colorful-square.jpg` | [rawpixel](https://www.rawpixel.com/image/5938538/colorful-houses-burano) |
| `city-brick-cobblestone-lane.jpg` | [StockSnap](https://stocksnap.io/photo/cobblestone-street-CURXEHYDN4) |
| `city-stone-arcade-colonnade.jpg` | [rawpixel](https://www.rawpixel.com/image/3297282/free-photo-image-old-hall-corridor-louvre-museum) |
| `city-rooftop-terrace-sunset.jpg` | [rawpixel](https://www.rawpixel.com/image/3299864/free-photo-image-city-rooftop-furniture) |
| `city-dutch-brick-lane.jpg` | [rawpixel](https://www.rawpixel.com/image/3291346/free-photo-image-alley-alleyway-arbour) |
| `city-whitewashed-andalusian-alley.jpg` | [StockSnap](https://stocksnap.io/photo/street-architecture-YALETRSLHN) |
| `city-oia-caldera-domes.jpg` | [rawpixel](https://www.rawpixel.com/image/5924824/photo-image-public-domain-blue-ocean) |
| `city-ghost-town-church.jpg` | [rawpixel](https://www.rawpixel.com/image/434105/free-photo-image-architecture-ghost-town-carol-highsmith) |
| `city-palace-fine-arts.jpg` | [StockSnap](https://stocksnap.io/photo/architecture-buildings-2DA5SJT6TT) |
| `city-parking-garage-sunset.jpg` | [StockSnap](https://stocksnap.io/photo/architecture-buildings-W1VNXBKR9U) |
| `city-bougainvillea-red-door.jpg` | [rawpixel](https://www.rawpixel.com/image/5971660/photo-image-wooden-wall-building) |
| `city-glass-facade-sidewalk.jpg` | [StockSnap](https://stocksnap.io/photo/city-street-CEX4LAQDOI) |
| `city-colorful-beach-huts.jpg` | [StockSnap](https://stocksnap.io/photo/houses-miniature-FXT5LMDCVE) |
| `city-sunset-terrace-sea.jpg` | [StockSnap](https://stocksnap.io/photo/santorini-greece-CA79EO5N7U) |
| `city-brick-townhouse-entrance.jpg` | [StockSnap](https://stocksnap.io/photo/building-facades-BAXJHTLAWN) |
| `city-seaside-promenade-sunset.jpg` | [StockSnap](https://stocksnap.io/photo/sea-pier-L3VD59EVDF) |
| `city-art-alley-lights.jpg` | [StockSnap](https://stocksnap.io/photo/alley-city-W98SRYZR3U) |
| `city-modern-office-walkway.jpg` | [rawpixel](https://www.rawpixel.com/image/6040191/photo-image-background-public-domain-construction) |
| `city-santorini-blue-domes.jpg` | [rawpixel](https://www.rawpixel.com/image/5905935/photo-image-background-public-domain-blue) |
| `city-venice-canal-night.jpg` | [rawpixel](https://www.rawpixel.com/image/3302235/free-photo-image-canal-boat-bridge) |
| `city-whitewashed-alley.jpg` | [StockSnap](https://stocksnap.io/photo/oia-walkway-F60ED0133D) |
| `city-chevron-wooden-door.jpg` | [rawpixel](https://www.rawpixel.com/image/6025881/wooden-door-free-public-domain-cc0-photo) |
| `city-riverside-tree-path.jpg` | [StockSnap](https://stocksnap.io/photo/city-park-RWOXTKPCD6) |
| `city-orange-modern-building.jpg` | [rawpixel](https://www.rawpixel.com/image/6034716/photo-image-public-domain-construction-free) |
| `city-belem-tower-pier.jpg` | [StockSnap](https://stocksnap.io/photo/tower-castle-DJ9G2OU1TO) |
| `city-whitewashed-village-alley.jpg` | [rawpixel](https://www.rawpixel.com/image/6034193/old-town-spain-free-public-domain-cc0-photo) |
| `city-brick-lighthouse-lawn.jpg` | [rawpixel](https://www.rawpixel.com/image/3090722/free-photo-image-america-architecture-block-island-light) |
| `city-arched-wooden-door.jpg` | [rawpixel](https://www.rawpixel.com/image/6038809/photo-image-public-domain-wood-house) |
| `city-night-outdoor-stairs.jpg` | [rawpixel](https://www.rawpixel.com/image/6032337/staircase-night-free-public-domain-cc0-image) |
| `city-blue-dome-bougainvillea.jpg` | [rawpixel](https://www.rawpixel.com/image/5922707/photo-image-flowers-public-domain-blue) |
| `city-rainbow-glass-building.jpg` | [rawpixel](https://www.rawpixel.com/image/6035012/photo-image-public-domain-wall-window) |
| `nature-autumn-path.jpg` | [rawpixel](https://www.rawpixel.com/image/6040003/photo-image-background-aesthetic-public-domain) |
| `nature-lavender.jpg` | [StockSnap](https://stocksnap.io/photo/flower-field-JVBWZNCQLW) |
| `nature-forest.jpg` | [rawpixel](https://www.rawpixel.com/image/8809263/photo-image-plant-light-tree) |
| `nature-snow-mountain.jpg` | [StockSnap](https://stocksnap.io/photo/snow-mountain-0KHG3WOP1F) |
| `nature-rocky-beach.jpg` | [StockSnap](https://stocksnap.io/photo/beach-ocean-VV40UIW8KF) |
| `nature-desert-dunes-grass.jpg` | [rawpixel](https://www.rawpixel.com/image/5904950/photo-image-background-public-domain-nature) |
| `nature-golden-beach-sunset.jpg` | [rawpixel](https://www.rawpixel.com/image/5907885/image-background-aesthetic-public-domain) |
| `nature-autumn-forest-path.jpg` | [rawpixel](https://www.rawpixel.com/image/11189509/nature-forest-path-woods-walkway) |
| `nature-misty-vineyard-dawn.jpg` | [StockSnap](https://stocksnap.io/photo/farm-field-CM4WVU6WO0) |
| `nature-alpine-lake-bridge.jpg` | [StockSnap](https://stocksnap.io/photo/mountain-water-4RH8BGXSVJ) |
| `nature-misty-moss-forest.jpg` | [rawpixel](https://www.rawpixel.com/image/5971806/row-trees-with-green-moss-the-forest) |
| `nature-alpine-lake-shore.jpg` | [StockSnap](https://stocksnap.io/photo/lake-lakdeside-WFJLWXPXYP) |
| `nature-misty-forest-path.jpg` | [StockSnap](https://stocksnap.io/photo/forest-hiking-TOKZSYWLKA) |
| `nature-tree-lined-path.jpg` | [StockSnap](https://stocksnap.io/photo/park-trees-1POHVCH6RG) |
| `nature-rippled-sand-beach.jpg` | [StockSnap](https://stocksnap.io/photo/beach-summer-0M1SPKYTR4) |
| `nature-red-dunes-mountains.jpg` | [rawpixel](https://www.rawpixel.com/image/5917297/image-public-domain-free-landscape) |
| `nature-iceland-mountain-road.jpg` | [StockSnap](https://stocksnap.io/photo/road-highway-GW8GYBILGU) |
| `nature-wooden-pier-rope.jpg` | [rawpixel](https://www.rawpixel.com/image/3295561/free-photo-image-pier-banister-bay) |
| `nature-lake-wooden-dock.jpg` | [StockSnap](https://stocksnap.io/photo/dock-lake-XQEEEWL3SO) |
| `nature-palm-white-beach.jpg` | [rawpixel](https://www.rawpixel.com/image/5925189/photo-image-background-clouds-public-domain) |
| `nature-redwood-forest-road.jpg` | [StockSnap](https://stocksnap.io/photo/rural-road-C8K19BYLLW) |
| `nature-lakeside-boathouse-dolomites.jpg` | [StockSnap](https://stocksnap.io/photo/mountain-valley-51XDMH7WJS) |
| `nature-palm-framed-beach.jpg` | [rawpixel](https://www.rawpixel.com/image/6036364/photo-image-public-domain-tree-nature) |
| `nature-glacial-lake-peaks.jpg` | [StockSnap](https://stocksnap.io/photo/lake-mountains-VZBJUVPO25) |
| `nature-turquoise-river-gorge.jpg` | [StockSnap](https://stocksnap.io/photo/valley-mountains-HIFDXE8C1V) |
| `nature-eucalyptus-gravel-path.jpg` | [StockSnap](https://stocksnap.io/photo/rural-road-IHS0A4DWT4) |
| `nature-jetty-twilight-lake.jpg` | [rawpixel](https://www.rawpixel.com/image/3305438/free-photo-image-pier-dock-boardwalk) |
| `nature-dune-grass-path.jpg` | [StockSnap](https://stocksnap.io/photo/beach-sand-2PT5PZPXLH) |
| `nature-aspen-lake-path.jpg` | [rawpixel](https://www.rawpixel.com/image/5965253/autumn-park-leaf-fall-lake) |
| `nature-palm-sand-beach.jpg` | [rawpixel](https://www.rawpixel.com/image/6020912/photo-image-background-aesthetic-public-domain) |
| `nature-misty-lake-bench.jpg` | [rawpixel](https://www.rawpixel.com/image/3282481/free-photo-image-nature-background-leaves) |
| `nature-dune-path-sea.jpg` | [StockSnap](https://stocksnap.io/photo/beach-sand-VR4V2BQROC) |
| `nature-turquoise-mountain-lake.jpg` | [rawpixel](https://www.rawpixel.com/image/3303528/free-photo-image-lake-louise-canada-nature) |
| `nature-glacial-river-valley.jpg` | [StockSnap](https://stocksnap.io/photo/mountain-valley-DLMN8FFBD9) |
| `nature-lupine-meadow-volcano.jpg` | [StockSnap](https://stocksnap.io/photo/lupine-flowers-9KMDQFRDE7) |
| `nature-forest-wooden-bridge.jpg` | [StockSnap](https://stocksnap.io/photo/forest-bridge-LPIPR9WFWV) |
| `nature-lupine-hillside-dusk.jpg` | [StockSnap](https://stocksnap.io/photo/lupine-flowers-CY6AJJTEV0) |
| `nature-bamboo-grove-path.jpg` | [rawpixel](https://www.rawpixel.com/image/6041837/photo-image-background-public-domain-plant) |
| `nature-formal-fountain-garden.jpg` | [rawpixel](https://www.rawpixel.com/image/3092128/free-photo-image-garden-array-america) |
| `nature-snowy-spruce-ridge.jpg` | [rawpixel](https://www.rawpixel.com/image/6029701/photo-image-public-domain-tree-forest) |
| `nature-alpine-lake-reflection.jpg` | [StockSnap](https://stocksnap.io/photo/mountain-valley-6SGPCZTXKL) |
| `nature-provence-lavender-rows.jpg` | [rawpixel](https://www.rawpixel.com/image/6035687/photo-image-background-flower-public-domain) |
| `nature-snowy-forest-clearing.jpg` | [rawpixel](https://www.rawpixel.com/image/6024161/photo-image-public-domain-tree-forest) |
| `nature-sea-cave-beach.jpg` | [rawpixel](https://www.rawpixel.com/image/2347508/photo-image-plant-tree-sky) |
| `nature-kirkjufell-waterfall.jpg` | [rawpixel](https://www.rawpixel.com/image/3300174/free-photo-image-waterfall-pool-beach) |
| `nature-white-rail-pier.jpg` | [rawpixel](https://www.rawpixel.com/image/3282450/free-photo-image-bridge-sea-banister) |
| `nature-mountain-forest-highway.jpg` | [rawpixel](https://www.rawpixel.com/image/3283474/free-photo-image-asphalt-cc0-countryside) |
| `nature-highland-glen-road.jpg` | [rawpixel](https://www.rawpixel.com/image/6032649/photo-image-background-clouds-public-domain) |
| `nature-bamboo-boardwalk-forest.jpg` | [rawpixel](https://www.rawpixel.com/image/5918019/image-public-domain-leaves-nature) |
| `nature-rapeseed-country-lane.jpg` | [StockSnap](https://stocksnap.io/photo/rural-road-TZ9ZLRTI63) |
| `nature-alpine-meadow-sunrise.jpg` | [StockSnap](https://stocksnap.io/photo/sunset-landscape-N7WJCFTERE) |
| `nature-snowy-mountain-trail.jpg` | [StockSnap](https://stocksnap.io/photo/switzerland-stoos-MOHYBQTEG4) |
| `nature-sunbeam-grove-meadow.jpg` | [StockSnap](https://stocksnap.io/photo/forest-tree-DFVOPIB0BN) |
| `nature-sunrise-beach-gulls.jpg` | [StockSnap](https://stocksnap.io/photo/nature-beach-DI64TAJTIS) |
| `nature-river-bluff-dusk.jpg` | [StockSnap](https://stocksnap.io/photo/canyon-river-H1RP4UZUTA) |
| `nature-dusk-lake-pier.jpg` | [rawpixel](https://www.rawpixel.com/image/5905708/dark-lake-view-free-public-domain-cc0-photo) |
| `nature-misty-rocky-slope.jpg` | [StockSnap](https://stocksnap.io/photo/trees-forest-20F265AAF9) |
| `nature-turquoise-sea-cliff.jpg` | [rawpixel](https://www.rawpixel.com/image/3297088/free-photo-image-blue-sky-coast-caribbean) |
| `nature-misty-larch-forest.jpg` | [StockSnap](https://stocksnap.io/photo/grey-dark-07DB6F371B) |
| `nature-lighthouse-grassy-dusk.jpg` | [StockSnap](https://stocksnap.io/photo/lighthouse-blue-B431BAFCE7) |
| `nature-alpine-meadow-peak.jpg` | [StockSnap](https://stocksnap.io/photo/stanserhorn-pilatus-J2FTIRCLY3) |
| `nature-sunlit-pine-forest.jpg` | [rawpixel](https://www.rawpixel.com/image/5939836/pine-forest) |
| `nature-misty-meadow-stream.jpg` | [rawpixel](https://www.rawpixel.com/image/3302252/free-photo-image-building-cc0-creative-commons) |
| `nature-lake-pukaki-road.jpg` | [rawpixel](https://www.rawpixel.com/image/4023111/alpine-landscape) |
| `nature-mossy-waterfall-forest.jpg` | [StockSnap](https://stocksnap.io/photo/nature-waterfalls-0WTZ2XPOJ1) |
| `nature-steel-truss-bridge.jpg` | [StockSnap](https://stocksnap.io/photo/bridge-structure-5Z56VEKLQY) |
| `nature-monument-valley-desert.jpg` | [rawpixel](https://www.rawpixel.com/image/5904983/photo-image-public-domain-mountain-free) |
| `nature-mountain-highway-forest.jpg` | [StockSnap](https://stocksnap.io/photo/landscape-rural-UVOVL6388T) |
| `nature-turquoise-tiered-falls.jpg` | [rawpixel](https://www.rawpixel.com/image/5903470/photo-image-light-public-domain-trees) |
| `nature-island-lighthouse-rocks.jpg` | [rawpixel](https://www.rawpixel.com/image/2372516/photo-image-public-domain-america-lighthouses) |
| `nature-misty-rocky-beach.jpg` | [StockSnap](https://stocksnap.io/photo/beach-sand-0509DJAHF2) |
| `nature-alpine-pine-valley.jpg` | [rawpixel](https://www.rawpixel.com/image/432772/free-photo-image-water-hills-cc0) |
| `nature-rocky-sea-cliffs.jpg` | [rawpixel](https://www.rawpixel.com/image/3293478/free-photo-image-beach-images-pictures-cc0-cliff) |
| `nature-palm-shade-beach.jpg` | [rawpixel](https://www.rawpixel.com/image/5905619/photo-image-public-domain-ocean-nature) |
| `nature-foggy-cliff-lighthouse.jpg` | [rawpixel](https://www.rawpixel.com/image/3286534/free-photo-image-lighthouse-seascape-building) |
| `nature-desert-road-dusk.jpg` | [StockSnap](https://stocksnap.io/photo/road-sunset-QVFEQ4HZBY) |
| `nature-poppy-field-mountains.jpg` | [rawpixel](https://www.rawpixel.com/image/3303196/free-photo-image-flower-field-meadow-poppy) |
| `nature-icelandic-waterfall-meadow.jpg` | [rawpixel](https://www.rawpixel.com/image/5975345/photo-image-public-domain-ocean-nature) |

## Your responsibility

What you generate with this app is yours to answer for: the galleries' terms above, the terms of the image API you configure, and the law where you are. Do not use it to depict or impersonate real people without their consent, or to depict minors.
