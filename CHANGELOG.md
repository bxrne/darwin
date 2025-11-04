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
