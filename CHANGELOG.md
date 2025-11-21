## [1.9.0](https://github.com/bxrne/darwin/compare/v1.8.1...v1.9.0) (2025-11-21)


### Features

* **config:** Add selection style config ([65164f9](https://github.com/bxrne/darwin/commit/65164f9bc71ecd72f3f0e1753dc57c387422d356))


### Bug Fixes

* **evo:** Fix evolution logic ([9e0b8ab](https://github.com/bxrne/darwin/commit/9e0b8ab9b5535db46a71b3ddc072a1c8d6c78cfa))

## [1.8.1](https://github.com/bxrne/darwin/compare/v1.8.0...v1.8.1) (2025-11-18)


### Bug Fixes

* **individual:** Fix recursion ([2f2626e](https://github.com/bxrne/darwin/commit/2f2626e73cebf01f6edc452bf9b755ab99678923))
* **test:** FIx test function calls ([32ea855](https://github.com/bxrne/darwin/commit/32ea855178badd362ba8d5cf7174073170f77d32))
* **test:** Linting is stupid ([ed537db](https://github.com/bxrne/darwin/commit/ed537db8548d5c8d47f80a297b5d50b160b08ed9))


### Miscellaneous Chores

* **ckpt:** hmm many questions ([d741745](https://github.com/bxrne/darwin/commit/d74174524c39f2b923b5804e78cd50c313fac24a))

## [1.8.0](https://github.com/bxrne/darwin/compare/v1.7.0...v1.8.0) (2025-11-14)


### Features

* **individual:** Ramped half and half ([59fb1e3](https://github.com/bxrne/darwin/commit/59fb1e318b117962ddbafb309ff39abf366f7e8c))


### Miscellaneous Chores

* **cmd:** Print out best formula ([70611c6](https://github.com/bxrne/darwin/commit/70611c606387e19b59dbd71ed50e0d0cef0a49eb))

## [1.7.0](https://github.com/bxrne/darwin/compare/v1.6.0...v1.7.0) (2025-11-11)


### Features

* **btree:** Add max depth constraint ([cdf935f](https://github.com/bxrne/darwin/commit/cdf935f4a81218260882094fd60df8e64af32815))


### Bug Fixes

* **tests:** Make tests run with new contract ([27f87e6](https://github.com/bxrne/darwin/commit/27f87e6bd375eee830dc6505ff0797323b3a07e2))

## [1.6.0](https://github.com/bxrne/darwin/compare/v1.5.0...v1.6.0) (2025-11-11)


### Features

* **plot:** plot experiments ([46f0929](https://github.com/bxrne/darwin/commit/46f0929b5bb65ae2012421d358bb7ffa37ccbc89))


### Miscellaneous Chores

* **plot:** Added plotting tool ([c15084e](https://github.com/bxrne/darwin/commit/c15084ee41ea039195d2ad8c967d652b243960d0))

## [1.5.0](https://github.com/bxrne/darwin/compare/v1.4.0...v1.5.0) (2025-11-11)


### Features

* **metrics:** Add csv output (configurable) of metrics ([b09efb9](https://github.com/bxrne/darwin/commit/b09efb9d779859e5e216551caa532a6b33da2a0b))

## [1.4.0](https://github.com/bxrne/darwin/compare/v1.3.0...v1.4.0) (2025-11-11)


### Features

* **tree:** Add correct crossover and apply penalty for /0 ([74a3278](https://github.com/bxrne/darwin/commit/74a3278f1862105e362e2ad872521e87d93e6dfa))
* **tree:** Add tree crossover ([08b705e](https://github.com/bxrne/darwin/commit/08b705eb905a53f8d09c5bdd13d6be875edb7d6b))


### Bug Fixes

* **btree, roulette:** Make roulette would with - and btree test pass ([707f385](https://github.com/bxrne/darwin/commit/707f38577a5de45f2de12a5d83ff5330638751a2))
* **btree:** Fix crossover bug ([0d23c54](https://github.com/bxrne/darwin/commit/0d23c5413b990f1fe5d9caa7c8389f80eb335b73))

## [1.3.0](https://github.com/bxrne/darwin/compare/v1.2.1...v1.3.0) (2025-11-11)


### Features

* **individual, evolution, cmd:** Add depth and description metrics to the metrics ([b048648](https://github.com/bxrne/darwin/commit/b048648986267e16a9da99729d6561528d9499c8))


### Bug Fixes

* **individual:** describe trees by using nice expression formatting ([f9534bb](https://github.com/bxrne/darwin/commit/f9534bbb11ad5962627d4bb9a07af74f429b3fbd))


### Miscellaneous Chores

* **evolution:** Remove old commented out code ([9447314](https://github.com/bxrne/darwin/commit/9447314ea17dfbcaa12ee20a51d8b4d9d2140179))

## [1.2.1](https://github.com/bxrne/darwin/compare/v1.2.0...v1.2.1) (2025-11-07)


### Bug Fixes

* **cmd:** added metrics done channel and waits for it before stats ([2a31726](https://github.com/bxrne/darwin/commit/2a317269002fc8bc9afd2be70a58e82da97c23b9))


### Tests

* **cmd:** receive extra return value from runEvolution ([17f2742](https://github.com/bxrne/darwin/commit/17f2742a927bd623c2a1b49e92e2435163537b8f))

## [1.2.0](https://github.com/bxrne/darwin/compare/v1.1.0...v1.2.0) (2025-11-07)


### Features

* **binaryInd:** Move binary individual fitness calculation to fitness.go ([cada853](https://github.com/bxrne/darwin/commit/cada853843062130018681e677b047aab0b09124))
* **evolution, individual, cmd:** added node-wise mutation to tree individual ([d902bd6](https://github.com/bxrne/darwin/commit/d902bd6cc1bc7c122f622c04c39fa9338529c175))
* **fitness:** Integrate new fitness func with bitgenome ([03037dd](https://github.com/bxrne/darwin/commit/03037dd3f3af607429ac34af7e7b3ee89d87b61c))
* **tree:** Add tree fitness evaluation geneation ([902e272](https://github.com/bxrne/darwin/commit/902e27243bbf5ceca0a53da8335fb51f9aef83f8))


### Bug Fixes

* **cmd, cfg:** Use primtive set in config for better creation of tree ([464ae60](https://github.com/bxrne/darwin/commit/464ae603073a6a5f52a474a87f243fa0390a1d93))
* **evolution, individual:** Use safe RNG, use clones of parents for modify safety ([fb99ca8](https://github.com/bxrne/darwin/commit/fb99ca8956b8983c7036f2b32e1ffafabb1cf86f))
* **release-changes.yml:** Only release linux bin ([72632e2](https://github.com/bxrne/darwin/commit/72632e26794677233f269f763baceca63b1c57c6))
* **test:** Remove unneeded tests and add binary fitness tests ([0743e09](https://github.com/bxrne/darwin/commit/0743e093234edbe5d71f1c76803dd1a6312bc831))
* **tree:** resolve merge conflicts ([4c59c2d](https://github.com/bxrne/darwin/commit/4c59c2dbf650cc5710fac2a848fd05ef440b81e2))

## [1.1.0](https://github.com/bxrne/darwin/compare/v1.0.1...v1.1.0) (2025-11-04)


### Features

* **cfg:** cfg can choose individual, cfg tests reduced in size based on testdata, table driven benchmark for each type ([50665b5](https://github.com/bxrne/darwin/commit/50665b5e1e3b30501d8586efeefd345dc4fb16eb))


### Bug Fixes

* **file:** Fix merge conflicts ([cc6f727](https://github.com/bxrne/darwin/commit/cc6f727125ce414d6fea993feff455e32a44fa61))
* **individual, README:** Tidy readme remove old rand lib usage ([325be84](https://github.com/bxrne/darwin/commit/325be84bfa47b6de70fd32453bfa2b0d46d367bf))
* **rng:** proper rng with new dep ([735992d](https://github.com/bxrne/darwin/commit/735992d0db45cae02aef0e6f499bce3ff1decbbd))


### Miscellaneous Chores

* **binary_individual:** better file name ([e8ca407](https://github.com/bxrne/darwin/commit/e8ca4075e725a13c5d5bc72913d9ff1180a8ea09))
* **ci:** move all tests to one place and only run on PR/Push ([764c705](https://github.com/bxrne/darwin/commit/764c705f5d8c2e12fb0a3d35732e6f0be270cdfa))
* **deps:** update ([694bd7f](https://github.com/bxrne/darwin/commit/694bd7fae40f2f2ca72614c1e202bfb98b3baf11))

## [1.0.1](https://github.com/bxrne/darwin/compare/v1.0.0...v1.0.1) (2025-11-04)


### Bug Fixes

* **rng:** fix rng to use a pool of a safe RNG ([101b2b9](https://github.com/bxrne/darwin/commit/101b2b9bc62fef7c51eef0c621ed8685c49e0f48))

## 1.0.0 (2025-11-04)


### Features

* **binary_tree:** add tree based stub for GP ([2fe4e07](https://github.com/bxrne/darwin/commit/2fe4e07d9cfea075f2bf55878fc0dec03f0cdd86))
* **cmd, config:** Added different configs and benchmarker ([de8ea6c](https://github.com/bxrne/darwin/commit/de8ea6c10f7249891f18aadbed71f9d2a47da3fd))
* **cmd, internal:** evo engine is now channel based, changed genome to bytes, added async metrics streaming ([cdba735](https://github.com/bxrne/darwin/commit/cdba73594a74ce8ef873f525b127e397fde29125))
* **EVA:** Add roulette selection and fix genome generation ([237e15f](https://github.com/bxrne/darwin/commit/237e15f5aa8f2a392cdf1b199ebfed45897b4188))
* **garden:** Add evolvable inteface to struct ([fd61ebb](https://github.com/bxrne/darwin/commit/fd61ebbd7077b020c52591a5d575068d1d778304))


### Bug Fixes

* **cfg, garden:** encap'd cfg validation in acquisition, added metrics to population ([958173f](https://github.com/bxrne/darwin/commit/958173f106661b22215ad94576d59f44463aa1b3))
* **cmd:** OOPs ([42800ca](https://github.com/bxrne/darwin/commit/42800ca4e4e1c77d324d08f574044ef60f5607ad))
* **evolution:** parallel'd build pop ([2bee5d0](https://github.com/bxrne/darwin/commit/2bee5d027ab7eaeb48a21884e554c1494ccc3929))
* **evolution:** param the creation of individual in popbuilder for swapping poptypes ([e91383d](https://github.com/bxrne/darwin/commit/e91383d6ae3a673e7d5a6ff4210d27fc611a03b4))
* **individual:** Fix mutate tests and rename tests ([f25f17d](https://github.com/bxrne/darwin/commit/f25f17d623993f125bb0a9cd4545f1745ea3537f))
* **metrics, tests:** fixed metric streamer stop and added tests ([906cb3a](https://github.com/bxrne/darwin/commit/906cb3aa8126a09074be454a774b40f76b58d2fa))
* **metrics:** protection from data races ([368aed9](https://github.com/bxrne/darwin/commit/368aed9673022cdcdede72c0b738e75d1917d76a))
* **mutate:** Fix mutate function and alter evolution contract ([a7b4b03](https://github.com/bxrne/darwin/commit/a7b4b03de9c1c408db473ddc1d2dae2bf4b7c590))
* **mutate:** Fix mutate function call ([77f945c](https://github.com/bxrne/darwin/commit/77f945c554b0690cb97dfdd52ecdcc920698ce31))
* **release-changes:** Update builder ([ebfaa45](https://github.com/bxrne/darwin/commit/ebfaa456718fde78217e1fb2372976c537ae4cc3))
* removed unused import ([8d99976](https://github.com/bxrne/darwin/commit/8d9997668b8c2716de1ee578937556e4b407b272))
* **rng, cfg, selection:** Added thread safe RNG with seeding ([4e75af4](https://github.com/bxrne/darwin/commit/4e75af4030bd402a87579f2b69695515460e0998))


### Code Refactoring

* **cmd, config, internal:** Added config file, package domain into garden ([562fc1c](https://github.com/bxrne/darwin/commit/562fc1c6d1bc061da8e7bf572ba4237d031755ef))
* **cmd:** encapsualted evolution main loop logic from main and bench ([78212ba](https://github.com/bxrne/darwin/commit/78212baaadb15491fd2743f5c4ee28db602875c5))


### Miscellaneous Chores

* **binaryIndividual:** remove duplicate file ([d2da913](https://github.com/bxrne/darwin/commit/d2da913d914b932df1fa787a44372b48fe17a17d))
* **ci, cmd, internal, pkg, testdata:** Added built,test,lint,vet CI and template go module ([cb9081b](https://github.com/bxrne/darwin/commit/cb9081b13e122d537eaa2223796a474b6dc7ac8a))
* **cmd:** added basic idea of GA ([b1b6321](https://github.com/bxrne/darwin/commit/b1b6321b0012bdc7a9ae9440e8cf563bc053ae96))
* **golangci:** back to v1 ([f1d0821](https://github.com/bxrne/darwin/commit/f1d0821a74fe5435c8e6268465b360b83bc2f17b))
* **golangci:** default to cfg ([72cb3dd](https://github.com/bxrne/darwin/commit/72cb3ddb05dca95c8bf923df50a0ea7696c9726b))
* **lint-vet:** fix download tag for lint ([069623a](https://github.com/bxrne/darwin/commit/069623ab67393b02fcebadb1aca51df365db1def))
* **lint-vet:** upgrade golangci-lint to v2 ([fff1ce5](https://github.com/bxrne/darwin/commit/fff1ce54a74a22e8044a89a91b88f6f0c5c8feb6))
* **release:** added semver changelogging ([5bb5e72](https://github.com/bxrne/darwin/commit/5bb5e7233468f7e7487ce3d84b373f4712d122de))


### Tests

* use proper test packages and remove unexported testcases ([575ab1b](https://github.com/bxrne/darwin/commit/575ab1b0040287a3fcde7f1216a42177a4aa744e))
