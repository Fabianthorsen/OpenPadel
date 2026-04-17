# Changelog

## [1.10.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.9.4...v1.10.0) (2026-04-17)


### Features

* **core:** SSE real-time updates + responsive drawer polish ([#41](https://github.com/Fabianthorsen/OpenPadel/issues/41)) ([ee27e3a](https://github.com/Fabianthorsen/OpenPadel/commit/ee27e3ab0c58f3047efde2e9297b8ec74c8e5806))
* **ui:** end tournament sheet menu ([#43](https://github.com/Fabianthorsen/OpenPadel/issues/43)) ([df42349](https://github.com/Fabianthorsen/OpenPadel/commit/df42349c77d3cc14648f962bdbe11c0b5696b6d2))


### Bug Fixes

* **auth:** admin access recovery via creator_user_id ([#42](https://github.com/Fabianthorsen/OpenPadel/issues/42)) ([d65545e](https://github.com/Fabianthorsen/OpenPadel/commit/d65545ec1c866ffc84cd8bafdaedaff8f4948dd2))
* **ui:** improve CreateDrawer styling and unify component patterns ([#39](https://github.com/Fabianthorsen/OpenPadel/issues/39)) ([a5475e0](https://github.com/Fabianthorsen/OpenPadel/commit/a5475e0e6cda10201a746e2efbbf46461ce873ee))

## [1.9.4](https://github.com/Fabianthorsen/OpenPadel/compare/v1.9.3...v1.9.4) (2026-04-14)


### Bug Fixes

* **mobile:** resolve iOS scroll lock and improve pull-to-refresh UX ([#36](https://github.com/Fabianthorsen/OpenPadel/issues/36)) ([c2c2ead](https://github.com/Fabianthorsen/OpenPadel/commit/c2c2eadca0a803423b51ec892cc29e6baf7613e8))

## [1.9.3](https://github.com/Fabianthorsen/OpenPadel/compare/v1.9.2...v1.9.3) (2026-04-14)


### Bug Fixes

* **mobile:** comprehensive PWA and mobile UX improvements ([#35](https://github.com/Fabianthorsen/OpenPadel/issues/35)) ([1f8e899](https://github.com/Fabianthorsen/OpenPadel/commit/1f8e899aa47c667d459e0ea9e9736a701dc12fa9))
* **mobile:** enable scrolling in PullToRefresh for lobby with many players ([#34](https://github.com/Fabianthorsen/OpenPadel/issues/34)) ([2b93866](https://github.com/Fabianthorsen/OpenPadel/commit/2b938666a0b24b50bc00dd0f59e5acae271cbc52))
* **mobile:** resolve P0 dialog and touch handling issues ([#32](https://github.com/Fabianthorsen/OpenPadel/issues/32)) ([05d5f4a](https://github.com/Fabianthorsen/OpenPadel/commit/05d5f4aa39f4e05ddfc2fcf38ae4a30f601d0d6a))

## [1.9.2](https://github.com/Fabianthorsen/OpenPadel/compare/v1.9.1...v1.9.2) (2026-04-13)


### Bug Fixes

* **ui:** improve PWA performance and touch handling on iOS ([a693812](https://github.com/Fabianthorsen/OpenPadel/commit/a6938126855eb0b0c09fc1a6a725592c0eed8113))

## [1.9.1](https://github.com/Fabianthorsen/OpenPadel/compare/v1.9.0...v1.9.1) (2026-04-13)


### Bug Fixes

* **ui:** move numpad to page level to fix viewport positioning ([7adb646](https://github.com/Fabianthorsen/OpenPadel/commit/7adb64654110639e753ecaf1e8bd28e2bf4f034e))

## [1.9.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.8.0...v1.9.0) (2026-04-13)


### Features

* **db:** add sqlc for type-safe query generation ([#26](https://github.com/Fabianthorsen/OpenPadel/issues/26)) ([023c24a](https://github.com/Fabianthorsen/OpenPadel/commit/023c24a1ceb7983403a575ed72848a0671337e73))
* **db:** add versioned migrations with goose ([#25](https://github.com/Fabianthorsen/OpenPadel/issues/25)) ([688768d](https://github.com/Fabianthorsen/OpenPadel/commit/688768d7d56d0d259e24094ea25de368b7a54495))

## [1.8.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.7.1...v1.8.0) (2026-04-12)


### Features

* **logging:** replace log.Printf with structured slog JSON logging ([665f14c](https://github.com/Fabianthorsen/OpenPadel/commit/665f14ce5d2d1b85e3aae240bb90d3e0e7b95143))
* **ui:** add toast notifications and internationalize API errors ([f980eca](https://github.com/Fabianthorsen/OpenPadel/commit/f980ecae15409bd5b4ed6dde380060d0032e9ee7))
* **ui:** pull-to-refresh on home, session, and profile screens ([5213f98](https://github.com/Fabianthorsen/OpenPadel/commit/5213f98d6ef9c262c908e47739f47cf270c0678e))


### Bug Fixes

* **lobby:** hide invite row when invited player has joined ([be56f59](https://github.com/Fabianthorsen/OpenPadel/commit/be56f59a4d21d5b02daec1739b50c72444d4cfe6))
* **logging:** skip logging successful GET requests ([9d0449b](https://github.com/Fabianthorsen/OpenPadel/commit/9d0449bfc75b422cec1f6e58853083d4cf02f0f1))


### Chores

* **roadmap:** mark v1.8.0 items done (pull-to-refresh, structured logging) ([d57f1ad](https://github.com/Fabianthorsen/OpenPadel/commit/d57f1ad538b3ae9495e7e4d2aee0b34ae4051cc3))

## [1.7.1](https://github.com/Fabianthorsen/OpenPadel/compare/v1.7.0...v1.7.1) (2026-04-10)


### Bug Fixes

* **deploy:** skip litestream restore if database already exists ([7609ec2](https://github.com/Fabianthorsen/OpenPadel/commit/7609ec262054586eb8bc7a923d8b105100a34a6d))

## [1.7.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.6.0...v1.7.0) (2026-04-09)


### Features

* **avatar:** add player avatar system with icon picker ([a08b5ff](https://github.com/Fabianthorsen/OpenPadel/commit/a08b5ff627dee5ae8b9ff12678b698106ec503a8))
* **ui:** avatar rings, players tab on-court/bench split, multi-court pending state ([707a7b0](https://github.com/Fabianthorsen/OpenPadel/commit/707a7b0aefcf4a8c04fcd54220698616b6f29814))
* **ui:** hide americano duration picker, sort upcoming tournaments by date ([d8de76c](https://github.com/Fabianthorsen/OpenPadel/commit/d8de76c54d15a99f096d3d4da790054cc44f6d9e))
* **ui:** redesign score screen with dark card, court tabs, numpad, and players tab ([3d55b55](https://github.com/Fabianthorsen/OpenPadel/commit/3d55b55579d8f74e0f902ae4d2dc666fef0edd1b))


### Bug Fixes

* **ui:** add court line pattern to score card background ([fbedab7](https://github.com/Fabianthorsen/OpenPadel/commit/fbedab75f22359280349999efc3a25e418e013f2))
* **ui:** add spacing between team score rows ([ea73b3c](https://github.com/Fabianthorsen/OpenPadel/commit/ea73b3c5790d9afafbe3adbf8ff896218177eac2))
* **ui:** constrain numpad bottom sheet to app max-width ([ee6419e](https://github.com/Fabianthorsen/OpenPadel/commit/ee6419e1371e5ef6f42f4f0aae97803c0a72ad12))
* **ui:** improve +/- button contrast on score card ([eb72a58](https://github.com/Fabianthorsen/OpenPadel/commit/eb72a58945bbe0ad69e3eb6abf96b1b0cbd522dc))
* **ui:** increase disabled button opacity from 30 to 40 ([5117154](https://github.com/Fabianthorsen/OpenPadel/commit/511715485806470231d89b61bbaca8fef6b7df45))
* **ui:** increase gap between team score rows ([7471c9a](https://github.com/Fabianthorsen/OpenPadel/commit/7471c9ac02c4264b5fef41bf24df3e1a5ea289ff))
* **ui:** large gap + white divider line between team score rows ([7b086b5](https://github.com/Fabianthorsen/OpenPadel/commit/7b086b5036f94d97a878c6f29ba015ab3b3e040f))
* **ui:** make court net line full opacity as team separator ([e2a9945](https://github.com/Fabianthorsen/OpenPadel/commit/e2a994512b1c70c6bf443deffdb68eefa19111b9))
* **ui:** make minus button visible with white border outline ([4e245ad](https://github.com/Fabianthorsen/OpenPadel/commit/4e245ad7bc49c3a9f87d7016984e9936d6eb595f))
* **ui:** match minus button style to plus button ([a5c5a3d](https://github.com/Fabianthorsen/OpenPadel/commit/a5c5a3d0ede9262987c50c324f4eb09a4a7c1da6))
* **ui:** remove NET pill divider from score card ([57e96b4](https://github.com/Fabianthorsen/OpenPadel/commit/57e96b45fe9f8543bf57abd840950df99fc3f1ce))
* **ui:** split score card into two separate team cards with gap ([301c785](https://github.com/Fabianthorsen/OpenPadel/commit/301c785dba981e872f0a41ef7977b3c2b298d00f))
* **ui:** use "Firstname L." format for player names on score card ([7f97fb5](https://github.com/Fabianthorsen/OpenPadel/commit/7f97fb5ea74638ba9aa6d9cd81d0593251208d62))
* **ui:** use lighter green for score card so avatars don't blend in ([be8c4f2](https://github.com/Fabianthorsen/OpenPadel/commit/be8c4f278a8a14eae1550ec7b74fffcc9006d2f4))


### Chores

* add CLAUDE.md, ROADMAP.md, and .claudeignore ([9d64a32](https://github.com/Fabianthorsen/OpenPadel/commit/9d64a320db1642c7dc3f32fe94194ed2af0f7656))
* add server binary to .gitignore ([8b45ff8](https://github.com/Fabianthorsen/OpenPadel/commit/8b45ff883130849e764aa048727f7c180a01c7f2))
* **ops:** add Litestream S3 replication, update Dockerfile, rewrite ARCHITECTURE.md, add UX specs ([7438132](https://github.com/Fabianthorsen/OpenPadel/commit/743813206da2f50707d6488da42fc811fc342c48))
* remove tracked server binary from repo ([d579ccd](https://github.com/Fabianthorsen/OpenPadel/commit/d579ccd55da814a70b2d922fd78f8ca918b5378c))
* require regression test before bug fixes in CLAUDE.md ([c261977](https://github.com/Fabianthorsen/OpenPadel/commit/c2619771d5553d601ac648d30893818679b1dcdb))
* update CLAUDE.md with no-direct-push rule and testing requirement ([4fcf70b](https://github.com/Fabianthorsen/OpenPadel/commit/4fcf70b3f637d8a3548fb463a038a4c829e66cf8))
* update ROADMAP.md after avatar system ([407a3cd](https://github.com/Fabianthorsen/OpenPadel/commit/407a3cd67a514583d05024c730a6b4f5cf65091e))

## [1.6.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.5.1...v1.6.0) (2026-04-08)


### Features

* **api:** add Mexicano game mode backend ([7d80b49](https://github.com/Fabianthorsen/OpenPadel/commit/7d80b4919e752a1f7c013f5a60bdda9e53cf9282))
* **mexicano:** add preset rounds option (4 / 6 / 8 / 10 / open) ([b70a326](https://github.com/Fabianthorsen/OpenPadel/commit/b70a326ff74daeaf529c15ba2cf45b5f5989f8c5))
* **mexicano:** enforce no-bench rule — exactly courts×4 players required ([9c8f605](https://github.com/Fabianthorsen/OpenPadel/commit/9c8f605edf7eb1b8e90a1c8ea9ce3b951a87bba8))
* **scheduler:** add Mexicano round generator with tests ([ef85980](https://github.com/Fabianthorsen/OpenPadel/commit/ef859806b286337ba9060577e275a9d07d43a27b))
* **scheduler:** randomize player order on tournament start ([de30b41](https://github.com/Fabianthorsen/OpenPadel/commit/de30b4112b45d465361da1dd30501e4304f8e0ca))
* **timer:** add court booking timer with rounds-or-time duration picker ([3bf8052](https://github.com/Fabianthorsen/OpenPadel/commit/3bf8052ac0bd1340621fc6af7add7e6f2a2bd84d))
* **ui:** add Mexicano mode to create flow and active session ([fcb6333](https://github.com/Fabianthorsen/OpenPadel/commit/fcb6333d288b19a1e83812a7a3b737b33c0aaeb7))
* **ui:** add rules info sheet to lobby for all game modes ([d70dfd2](https://github.com/Fabianthorsen/OpenPadel/commit/d70dfd2843e34176de6fd07c63f6bdf8cb338801))


### Bug Fixes

* **scheduler:** randomise Team A/B sides so admin isn't always Team A ([3a7a5e7](https://github.com/Fabianthorsen/OpenPadel/commit/3a7a5e7e2d535e27710964a36ec4b48ebe9a5034))
* **ui:** move rules info button to main lobby nav (visible to all joined players) ([12078fe](https://github.com/Fabianthorsen/OpenPadel/commit/12078fe62df52c4191cabac38797bda6ad2eb111))
* **ui:** show correct game mode name in invite/lobby title ([9e4f622](https://github.com/Fabianthorsen/OpenPadel/commit/9e4f622a83ffd4ecbb165350eab96b2bb2657e6d))
* **ui:** show round number in leaderboard for Mexicano (no fixed total) ([4c3533d](https://github.com/Fabianthorsen/OpenPadel/commit/4c3533d9b5cc1a1077358c677878fcffa81d908f))


### Chores

* remove dead code ([76bc178](https://github.com/Fabianthorsen/OpenPadel/commit/76bc1787ed6bc619f9e19fe10c8ba86cd4d31f2b))

## [1.5.1](https://github.com/Fabianthorsen/OpenPadel/compare/v1.5.0...v1.5.1) (2026-04-08)


### Bug Fixes

* **scheduler:** ensure minimum meaningful rounds for bench configs ([716cfdf](https://github.com/Fabianthorsen/OpenPadel/commit/716cfdf8f07b377ef5a7b8429113848245cc6110))

## [1.5.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.4.2...v1.5.0) (2026-04-08)


### Features

* **sessions:** add ended_early flag to track prematurely ended tournaments ([271c400](https://github.com/Fabianthorsen/OpenPadel/commit/271c400387609b46cc5edb718d96d3a404a9fe6f))
* **ui:** improve tournament end/cancel button styling and UX ([1be7ec7](https://github.com/Fabianthorsen/OpenPadel/commit/1be7ec7b00ca980f5cf9e0770051112f181bde2d))

## [1.4.2](https://github.com/Fabianthorsen/OpenPadel/compare/v1.4.1...v1.4.2) (2026-04-07)


### Bug Fixes

* **leaderboard:** add tiebreaker chain for fair ranking ([3317bb8](https://github.com/Fabianthorsen/OpenPadel/commit/3317bb846e97803930ce6f3d2f9c14ed3cdc5767))

## [1.4.1](https://github.com/Fabianthorsen/OpenPadel/compare/v1.4.0...v1.4.1) (2026-04-07)


### Bug Fixes

* **scheduler:** correct round count for bench configurations ([93c6ba6](https://github.com/Fabianthorsen/OpenPadel/commit/93c6ba67b863b0028cdbfd2cc64ad3a06f3b4ecd))

## [1.4.0](https://github.com/Fabianthorsen/OpenPadel/compare/v1.3.2...v1.4.0) (2026-04-06)


### Features

* 4-char uppercase join codes with improved display ([12e1171](https://github.com/Fabianthorsen/OpenPadel/commit/12e1171c12e6353bda87a8c9cb194cb2050304a1))
* adaptive polling — 3s in lobby, 15s during active play ([d927002](https://github.com/Fabianthorsen/OpenPadel/commit/d927002cb9123f7fc0240fb9b53763b855467faf))
* add session code entry on home page ([9aa47d5](https://github.com/Fabianthorsen/OpenPadel/commit/9aa47d507054d2aaef1acda0c481f7793b6cda18))
* admin joins as player with creator crown ([9d22ced](https://github.com/Fabianthorsen/OpenPadel/commit/9d22cedbab60f79c52870fd4b4ea95a3ffdc844d))
* allow admin to add players manually in lobby ([42c1e55](https://github.com/Fabianthorsen/OpenPadel/commit/42c1e5584edc354e539b6463ee73d7b89c483450))
* **auth:** email/password authentication with persistent user accounts ([07e0365](https://github.com/Fabianthorsen/OpenPadel/commit/07e036523663524436b2c9d0a8779aa0c24f73a9))
* **auth:** user accounts, profiles, tournament history, and password reset ([bec8e33](https://github.com/Fabianthorsen/OpenPadel/commit/bec8e33d08c16efc814f80b76fa903420be82bec))
* cancel tournament anytime, confirm dialogs, remove player in lobby ([6ddd1ae](https://github.com/Fabianthorsen/OpenPadel/commit/6ddd1aed0a836596c523fa63c00b31d9a55ff971))
* cancel tournament, fix minimum player count ([d6fd3d2](https://github.com/Fabianthorsen/OpenPadel/commit/d6fd3d2b1b2129a8745534848877df01d7e27a59))
* combine join code and share link into one card ([c06fba8](https://github.com/Fabianthorsen/OpenPadel/commit/c06fba8dc9052ed34a240ad7167053ebad6a8048))
* complete session shows final leaderboard with new session CTA ([aa3159d](https://github.com/Fabianthorsen/OpenPadel/commit/aa3159d5f33417519846f0093b82a5ff47142db7))
* **contacts:** add contact buttons to final results screen ([3d134ae](https://github.com/Fabianthorsen/OpenPadel/commit/3d134aeda9fb55883880e0077ef38d170d029a2a))
* **contacts:** add contacts feature with search and profile UI ([3b4288c](https://github.com/Fabianthorsen/OpenPadel/commit/3b4288c244637120ceba109555e43f1034ea8ece))
* **contacts:** add contacts picker to tournament creation drawer ([3aeac78](https://github.com/Fabianthorsen/OpenPadel/commit/3aeac78d32c6bff4cba928bee307bdc51cd9fd9b))
* deploy pipeline — Dockerfile, fly.toml, GitHub Actions ([e09ec7a](https://github.com/Fabianthorsen/OpenPadel/commit/e09ec7aad39cd2a3ff12a242c56b7f3a295a9437))
* explicit round advance, edit score fix, and draws support ([bd87642](https://github.com/Fabianthorsen/OpenPadel/commit/bd876425e4be8ea9e06da64cee71724c3b31f0b3))
* **i18n:** add English and Norwegian translations via svelte-i18n ([c9af593](https://github.com/Fabianthorsen/OpenPadel/commit/c9af59378a8f35bef1a945a5329d6e4872453258))
* invite system, improved lobby, new favicon ([531c620](https://github.com/Fabianthorsen/OpenPadel/commit/531c620e67dbf8916525fac9ede46d4c79271760))
* **invites:** add invite flow — contacts must accept before joining ([6344c50](https://github.com/Fabianthorsen/OpenPadel/commit/6344c507c4d03b669489544814b5306d129328b0))
* **leaderboard:** redesign standings with leader hero card and W/L stats ([41c0361](https://github.com/Fabianthorsen/OpenPadel/commit/41c0361ef0472a6196538eb34322ff9ff31d0812))
* live round view with score entry and leaderboard ([1431a7a](https://github.com/Fabianthorsen/OpenPadel/commit/1431a7aaaafb0214972e881e843c43fe62d7b569))
* **livescores:** real-time live score sync with in-memory store ([ac6674f](https://github.com/Fabianthorsen/OpenPadel/commit/ac6674fc71dcf014ff1f7374d5ff9b40a7c3351a))
* open score entry to all players, fix layout and design ([79b164b](https://github.com/Fabianthorsen/OpenPadel/commit/79b164b0158909da508532b2c8f491ca7927e289))
* **profile:** career stats page and pre-fill join form from user profile ([2a0b381](https://github.com/Fabianthorsen/OpenPadel/commit/2a0b3815e2350b436357cb8cb1084d8feb3d9a30))
* **profile:** split career stats by game mode; misc UI polish ([e9af34d](https://github.com/Fabianthorsen/OpenPadel/commit/e9af34dd74f4520cb4fda6f30cb1afebd081c752))
* **push:** web push notifications for tournament start ([16721d0](https://github.com/Fabianthorsen/OpenPadel/commit/16721d06f36c8cd0c5029773c07dafe456da91b5))
* redesign active session score view ([6266d52](https://github.com/Fabianthorsen/OpenPadel/commit/6266d5299ccf413c1e6ad1c7ae2f55bd78ae0523))
* restore PWA and add OG share tags for session invites ([3a0f3ca](https://github.com/Fabianthorsen/OpenPadel/commit/3a0f3ca5c8ed6f55cddf3299b722e2581e01a0a6))
* scaffold Go backend ([6338c45](https://github.com/Fabianthorsen/OpenPadel/commit/6338c456723e271e0c5d4357a79fc512b34d1666))
* scaffold SvelteKit frontend ([23c0c4f](https://github.com/Fabianthorsen/OpenPadel/commit/23c0c4fbd1e0e4c476fdfaec179bf53d916126fe))
* session setup screen and lobby with invite link ([b38a5d1](https://github.com/Fabianthorsen/OpenPadel/commit/b38a5d1d7c7c6c17f17d01d7c1b999274b1a16bb))
* shake + red ring on name input validation error ([91a3b56](https://github.com/Fabianthorsen/OpenPadel/commit/91a3b567d92b37475b1dafbfaa82d1ff70c2b8e7))
* show join code prominently in lobby ([95275c3](https://github.com/Fabianthorsen/OpenPadel/commit/95275c33d397e895d33987701261d15abd8d7875))
* **tennis:** regular 2v2 game mode with sets, serve tracking, and team assignment ([a2d02e7](https://github.com/Fabianthorsen/OpenPadel/commit/a2d02e70197bdfa1d5446274de35eec8858fd48b))
* **tournament:** add naming support and fun awards on final results ([1ef5966](https://github.com/Fabianthorsen/OpenPadel/commit/1ef5966663264dfc91e19f0dd0fac6538088995a))
* **ui:** create tournament drawer, session → tournament copy ([#2](https://github.com/Fabianthorsen/OpenPadel/issues/2)) ([f5df663](https://github.com/Fabianthorsen/OpenPadel/commit/f5df6631506720bb7e54390545cc53dc384eefb5))
* **ui:** final results screen with podium, rankings, and summary card ([b9f0dc0](https://github.com/Fabianthorsen/OpenPadel/commit/b9f0dc016bea89bd591ab70939b89dbcb020b3f4))
* **ui:** highlight match winner in scored court card ([8459733](https://github.com/Fabianthorsen/OpenPadel/commit/8459733c6ef89804cb5ed3227826b75797c6f394))
* **ui:** podium colours on leaderboard rows (1st forest green, 2nd mid, 3rd light) ([2cec52b](https://github.com/Fabianthorsen/OpenPadel/commit/2cec52bb40db66a1da31bf8415e32e3891e37146))
* **ui:** profile page redesign with collapsible sections and join by code ([80a9ff5](https://github.com/Fabianthorsen/OpenPadel/commit/80a9ff52530b001c4b5e1408cc73286ff5f15860))
* **ui:** tap result card to edit scores directly, hide next round while editing ([012bf3e](https://github.com/Fabianthorsen/OpenPadel/commit/012bf3e5792822a83a7836af62acb374f92a878e))
* **ui:** v1 polish — i18n, serve indicator, nav icons, rejoin ([d1d12f0](https://github.com/Fabianthorsen/OpenPadel/commit/d1d12f0afeb5a0ca4863c3647131178662ed9bfe))
* wire up shadcn-svelte Button and Input components ([bb2ff2e](https://github.com/Fabianthorsen/OpenPadel/commit/bb2ff2e089b9389776da4555faa1a38921202c46))


### Bug Fixes

* don't overwrite admin's player ID when adding players manually ([327b2bc](https://github.com/Fabianthorsen/OpenPadel/commit/327b2bc377d892f9b58fa73cb3102a5077b15b61))
* fly.toml http_service should be table not array ([4334d2e](https://github.com/Fabianthorsen/OpenPadel/commit/4334d2e03aa3c8a970bdf55fe7f11deea9830fce))
* immediate refresh on join, share fallback on desktop ([c79d25c](https://github.com/Fabianthorsen/OpenPadel/commit/c79d25c7dab91994ff217dddfd6ea68a867b361d))
* **profile:** show game mode in upcoming tournaments; reduce bottom padding ([31aae1f](https://github.com/Fabianthorsen/OpenPadel/commit/31aae1f261eae2e5905e368136f155a825a469f2))
* **push:** better error propagation and SW timeout; fix stats card alignment ([a4d48c7](https://github.com/Fabianthorsen/OpenPadel/commit/a4d48c7a3e6fd0b838c5c73f76d25aac830306d7))
* **pwa:** add apple-touch-icon for iOS home screen icon ([d4a2cc0](https://github.com/Fabianthorsen/OpenPadel/commit/d4a2cc0d4509f7aab7e02a10f223f866d33b4d57))
* **pwa:** switch to SvelteKit-native SW to fix push on iOS ([b6f0851](https://github.com/Fabianthorsen/OpenPadel/commit/b6f0851696ea8c8d22504d854af28406b67bf01a))
* redirect home on 404, handle 204 in API client ([3834f15](https://github.com/Fabianthorsen/OpenPadel/commit/3834f1518207d8f88f78eae20752e5979cdd5235))
* remove VitePWA to resolve _app/immutable redirect loop ([b325936](https://github.com/Fabianthorsen/OpenPadel/commit/b3259364848b35a169fcc6be070709829003f5d7))
* replace Svelte logo with NT monogram icon and remove debug logging ([6f8fe87](https://github.com/Fabianthorsen/OpenPadel/commit/6f8fe87efb8f2f4a2c45e72aefbc0cd66576539f))
* **scheduler:** replace rotation-based algorithm with full-history random search ([dcc241f](https://github.com/Fabianthorsen/OpenPadel/commit/dcc241f6371f2e8f0d49017d54454127902c464b))
* **scheduler:** rounds_total = N-1 players (correct Americano rotation) ([74096df](https://github.com/Fabianthorsen/OpenPadel/commit/74096df984f740ce490f6a25a20a5637afd95b12))
* serve index.html directly to prevent SPA redirect loop ([d9f9d42](https://github.com/Fabianthorsen/OpenPadel/commit/d9f9d42144c25f9eb71231ffa9a327f524aa281a))
* serve static assets via ServeContent, add request logging ([0c54fb4](https://github.com/Fabianthorsen/OpenPadel/commit/0c54fb48fa087de1088ae765588c51b05057c125))
* **sessions:** end tournament deletes session; remove reopen ([5bbd3b5](https://github.com/Fabianthorsen/OpenPadel/commit/5bbd3b58d1cbeb842dc48b5bbbdb534f37cf29f2))
* **ui:** center draw pill on divider line between team rows ([95a5e5c](https://github.com/Fabianthorsen/OpenPadel/commit/95a5e5c3768a591da9147c2a59aa6a0b8e57a687))
* **ui:** derive serve from score (rotates every 4 points in Americano) ([70d8d3f](https://github.com/Fabianthorsen/OpenPadel/commit/70d8d3f3b81a6a45787d5147ee9dba1f27b8e194))
* **ui:** disable pinch-to-zoom on iOS via touch-action CSS ([5bda4e0](https://github.com/Fabianthorsen/OpenPadel/commit/5bda4e0e1d85f92faa773c064864902128cc185a))
* **ui:** prevent double-tap zoom on interactive elements across all clients ([465d0e1](https://github.com/Fabianthorsen/OpenPadel/commit/465d0e1e9e9339bb2b2cf911a7825c13f8c76492))
* **ui:** rejoin href computed in script, replace ICU plural with ternary ([9d51e4b](https://github.com/Fabianthorsen/OpenPadel/commit/9d51e4bab9676604c88da594fe123eb0a8b46cf9))
* **ui:** remove redundant rank label from leader hero card ([3498a9f](https://github.com/Fabianthorsen/OpenPadel/commit/3498a9f9060a35b4be7e6503de96fda483e52ba3))
* **ui:** replace copy text with share/check icon on invite link ([c2f029f](https://github.com/Fabianthorsen/OpenPadel/commit/c2f029f6822688c7de11dad781c339610fef00dd))
* **ui:** restyle cancel and dev seed buttons in lobby ([901d658](https://github.com/Fabianthorsen/OpenPadel/commit/901d658e4ca2cb169ed060e1ec24f5cf5c769ef7))
* **ui:** show spinner when cancelling tournament instead of default screen ([bcf34a3](https://github.com/Fabianthorsen/OpenPadel/commit/bcf34a305182646e6d72393fd201565f95b905ba))
* **ui:** today date in calendar turns white when selected ([22e07aa](https://github.com/Fabianthorsen/OpenPadel/commit/22e07aa05b085e6e8890f61444452f660588eebb))
* **ui:** use current date for final results summary card ([9c260e3](https://github.com/Fabianthorsen/OpenPadel/commit/9c260e34fc4693ab220e785a93ec0d1ce0dacbbf))
* use all:build embed to include _app/ directory ([280b96e](https://github.com/Fabianthorsen/OpenPadel/commit/280b96e2ba3d77fbaae48bc80e5129e81b51fe1f))


### Chores

* add conventional commits and release-please ([a38c068](https://github.com/Fabianthorsen/OpenPadel/commit/a38c068baa2a129a3b4c5c10666bcd454893425d))
* **ci:** deploy on version tags only, not every commit to main ([54575c7](https://github.com/Fabianthorsen/OpenPadel/commit/54575c72281bd6b902c5eeca498e70ede361403e))
* initial project scaffold ([301aa05](https://github.com/Fabianthorsen/OpenPadel/commit/301aa051c71a1cf130a2f60b636a3acaf29a9f37))
* release v1.0.0 ([50f9e28](https://github.com/Fabianthorsen/OpenPadel/commit/50f9e283a3b7247be99e5572f8c371906cff5334))
* remove footer copyright text ([2c5502a](https://github.com/Fabianthorsen/OpenPadel/commit/2c5502a80fa69f06efe76c7bdb7c6a7c60052ae0))
* rename NotTennis → OpenPadel ([17808ed](https://github.com/Fabianthorsen/OpenPadel/commit/17808edc1528e03ffb8401b943b90176c0f075d7))
* sync release-please manifest to v1.3.2 ([b4c0f6c](https://github.com/Fabianthorsen/OpenPadel/commit/b4c0f6cca190a33c0c2bd90802e774bf9e943a62))
* **v1.2:** misc fixes, i18n, auth, store and API cleanup ([ed93fa0](https://github.com/Fabianthorsen/OpenPadel/commit/ed93fa074a6906674e9047d7ca37276eb3ad603b))
