## [1.20.0](https://github.com/bxrne/darwin/compare/v1.19.0...v1.20.0) (2025-12-10)


### Features

* **config:** add ge_problem configuration file and update tree fitness calculation to handle NaN/Inf results ([8dc4c15](https://github.com/bxrne/darwin/commit/8dc4c15aa61064104b0e0e8a55e54b12273bd741))


### Bug Fixes

* **evolution, fitness:** fix crossover and mutation probability in offspring generation; adjust penalty for NaN/Inf in fitness calculation ([7dbbca7](https://github.com/bxrne/darwin/commit/7dbbca704274500f524924aff77dc833cbf1cd37))


### Code Refactoring

* **fitness:** adjust penalty for NaN/Inf in fitness calculation and enhance genome description for logging ([b9bd51d](https://github.com/bxrne/darwin/commit/b9bd51d26379e0b437c1e15f145a05200707277c))
* **tree:** enhance Describe method for improved genome representation and readability ([ad188bb](https://github.com/bxrne/darwin/commit/ad188bb2726b22546d113ca7412711ca20bed27c))


### Miscellaneous Chores

* **config:** update ge_problem configuration parameters for evolution and fitness settings ([c6321d5](https://github.com/bxrne/darwin/commit/c6321d5a27bb04abffe184bbe1be7ae04a1690f1))

## [1.19.0](https://github.com/bxrne/darwin/compare/v1.18.0...v1.19.0) (2025-12-10)


### Features

* track winning client and change variables for training ([ff39f1b](https://github.com/bxrne/darwin/commit/ff39f1b0296b53671794d951a369e1cfa7859e52))


### Bug Fixes

* **cmd:** logging issue ([09cb5ee](https://github.com/bxrne/darwin/commit/09cb5eef2feccd83085030cd690e1a745fe777fa))


### Miscellaneous Chores

* **cmd, internal:** log noise reduced at info level ([744634c](https://github.com/bxrne/darwin/commit/744634c47d29b27e86f65122044ef8d0645d781d))
* fixup logging ([58afd43](https://github.com/bxrne/darwin/commit/58afd4311431720df47ef979168be7445b81d785))

## [1.18.0](https://github.com/bxrne/darwin/compare/v1.17.0...v1.18.0) (2025-12-08)


### Features

* **game, cfg:** implementing expander mode (on 20 cumul. wins global_ref) ([3e09d21](https://github.com/bxrne/darwin/commit/3e09d213f9768b251faa9343562521e282757267))

## [1.17.0](https://github.com/bxrne/darwin/compare/v1.16.1...v1.17.0) (2025-12-08)


### Features

* **cmd:** Add test cases to fitness facade (n games per individual) ([0d9ea6b](https://github.com/bxrne/darwin/commit/0d9ea6bd91e9d11621452c0d7e0d697a26870ad9))


### Bug Fixes

* **cmd, cfg, fitness:** use test_case_count games per individual (weight + tree) ([11c9111](https://github.com/bxrne/darwin/commit/11c91119e46e5e7e9481a2f947483f7cbe62810d))
* **fitness:** call to calc setup with TC count ([92fa2fe](https://github.com/bxrne/darwin/commit/92fa2fe4de6f67ea511455f41173b283c8f6568f))

## [1.16.1](https://github.com/bxrne/darwin/compare/v1.16.0...v1.16.1) (2025-12-08)


### Bug Fixes

* **game:** fix log formatting and remove unused files ([f605a99](https://github.com/bxrne/darwin/commit/f605a99317bf3945b29d2540158606aeade08389))

## [1.16.0](https://github.com/bxrne/darwin/compare/v1.15.0...v1.16.0) (2025-12-08)


### Features

* Better reward function and replays ([666bb1b](https://github.com/bxrne/darwin/commit/666bb1bf4de7dda98735f3855bd8fb45fafd4b41))


### Bug Fixes

* **cmd:** remove uneeded % in if ([e12c95d](https://github.com/bxrne/darwin/commit/e12c95d77c8c4da5a5f6cb8c991fe9d4322bed79))
* **fitness:** softmaxing the action sampling ([e3c5313](https://github.com/bxrne/darwin/commit/e3c5313e592c5689cbca4c5a8e6f4100812a1920))


### Miscellaneous Chores

* **game:** del smoketest ([5017846](https://github.com/bxrne/darwin/commit/5017846b17e818fbd0eff2afe865dafd7e2f5af6))

## [1.15.0](https://github.com/bxrne/darwin/compare/v1.14.1...v1.15.0) (2025-12-08)


### Features

* **actions:** Implement proper masking and update reward function to not punish a pass as an invlaid action(might need to change) ([ef2611b](https://github.com/bxrne/darwin/commit/ef2611bf3403420a1690e095e28a64bafd44063d))


### Bug Fixes

* **ati crossover:** Make ati use correct crossover ([12f159a](https://github.com/bxrne/darwin/commit/12f159ae0b38953c1cf2b761aef4dc1c9bb961e3))
* **tests:** Fix softmax tests and remove old validator tests ([4a57c72](https://github.com/bxrne/darwin/commit/4a57c72ee435dde91cc508549c6c2d08659aa5e5))

## [1.14.1](https://github.com/bxrne/darwin/compare/v1.14.0...v1.14.1) (2025-12-05)


### Bug Fixes

* **game:** remove conflict tags ([4b693c6](https://github.com/bxrne/darwin/commit/4b693c69796457c6bf17f8d23a2f41d311ae91e2))

## [1.14.0](https://github.com/bxrne/darwin/compare/v1.13.0...v1.14.0) (2025-12-05)


### Features

* Bridge fully working ([1108f9e](https://github.com/bxrne/darwin/commit/1108f9e2d32e4e99e3dc7bd31d09f184e31dc0e8))
* **bridge:** WIP Add procssors to allow concurrent games to bridge ([365285c](https://github.com/bxrne/darwin/commit/365285c2a0ff42caf3e19360b3dffd3b2274c1be))
* **game:** Game working from go to python loop ([57fe9c6](https://github.com/bxrne/darwin/commit/57fe9c61068b506dae119c693c7cfd6b846eed1a))
* **integration:** integrate code to run properly ([a8db93d](https://github.com/bxrne/darwin/commit/a8db93d08acf62b7a79fb1f0f36ff68f54b671f0))


### Bug Fixes

* **bridge:** Fix the bridge ([a2eaa04](https://github.com/bxrne/darwin/commit/a2eaa04e5bf18a66b0c66458d9631ae4f2fcb52b))
* **fitness:** Mask actions. mountain no-op and adjacency and bounds ([a785319](https://github.com/bxrne/darwin/commit/a785319275a39e6b05a905a13e9ba44bfd74c40a))
* **game, fitness:** better logging for game and fitness fixes ([55251bd](https://github.com/bxrne/darwin/commit/55251bd701a1fe6edaf2d09803639a1aaecfb6b0))
* **game, internal:** fix encodign nyumpy values ([88b1ca8](https://github.com/bxrne/darwin/commit/88b1ca8fdf1f5b5c4c69a8a34b6ca6f5029684bb))
* **game:** Add set map size and return bool arr for info ([bdfa669](https://github.com/bxrne/darwin/commit/bdfa669e14e193e66dc77fd1ba877fd8c476d2f1))
* **game:** owned cells marshalling ([a750ba3](https://github.com/bxrne/darwin/commit/a750ba385b2aa956a4038ca0a80b2ea0c7c1ee53))
* **game:** reduce logging noise and better cleanup ([1747e2d](https://github.com/bxrne/darwin/commit/1747e2d99ef765d4a925d1941a8fc730f3df949d))
* **lint:** Fix lint issues ([290048a](https://github.com/bxrne/darwin/commit/290048acf6fab92fdb30364d577fa63e608bdc7c))
* **population:** not waiting for fitness goroutines fixed ([42f8e8c](https://github.com/bxrne/darwin/commit/42f8e8c5b3e792cfcae5b18ec500bb8cdd578e0b))
* reuse tcp conns w pool, added healthcheck, better error handling ([b19510b](https://github.com/bxrne/darwin/commit/b19510b22618e5ac14fa19486c893bbd2373fa0f))


### Code Refactoring

* **cfg:** Change actions to tuple array to allow for action counts to be stored ([903c990](https://github.com/bxrne/darwin/commit/903c99000ae40e9a7518ce98af4767fe840746f3))


### Miscellaneous Chores

* **cfg, fitness:** update test expectations and signature fixes ([35d3f11](https://github.com/bxrne/darwin/commit/35d3f11955d6387cf491a29b8a595fd3c9e26227))
* **fitness:** remove unneeded func ([d060235](https://github.com/bxrne/darwin/commit/d06023567c117467e996df90112fbb87716bcdd9))

## [1.13.0](https://github.com/bxrne/darwin/compare/v1.12.4...v1.13.0) (2025-12-02)


### Features

* **game:** Use frequent asset reward fn in game ([357abdf](https://github.com/bxrne/darwin/commit/357abdfbd40280d4d5542703856672896683fe5d))

## [1.12.4](https://github.com/bxrne/darwin/compare/v1.12.3...v1.12.4) (2025-12-02)


### Bug Fixes

* fix callers to New ATI ([eb9d7a3](https://github.com/bxrne/darwin/commit/eb9d7a33c0bc766a322eebaa1dd0651112b5e902))

## [1.12.3](https://github.com/bxrne/darwin/compare/v1.12.2...v1.12.3) (2025-12-02)


### Bug Fixes

* **individual:** Ati create remove unneeded vars ([e9d7d47](https://github.com/bxrne/darwin/commit/e9d7d47a35c419febab3bc44eb28980b48ae89fa))

## [1.12.2](https://github.com/bxrne/darwin/compare/v1.12.1...v1.12.2) (2025-12-02)


### Reverts

* Revert "fix(individual): Culled params from new random ATI" ([b143aba](https://github.com/bxrne/darwin/commit/b143aba54617218caa65ae0af1eb5f3289dd02b3))
* Revert "fix(demo, individual): Fix calls to Random ATI init" ([22410da](https://github.com/bxrne/darwin/commit/22410da20e7260abf0ad2d18ed828c723bd9b056))

## [1.12.1](https://github.com/bxrne/darwin/compare/v1.12.0...v1.12.1) (2025-12-02)


### Bug Fixes

* **demo, individual:** Fix calls to Random ATI init ([bb3f692](https://github.com/bxrne/darwin/commit/bb3f692f65b895a54484f61c178c851dbdefefa4))
* **individual:** Culled params from new random ATI ([3f5bebd](https://github.com/bxrne/darwin/commit/3f5bebd4fa7175e293fc0f697c183f53dd19a96d))

## [1.12.0](https://github.com/bxrne/darwin/compare/v1.11.0...v1.12.0) (2025-12-02)


### Features

* **plumbing:** Basic plumbing implementation for two populations ([7e5a9c4](https://github.com/bxrne/darwin/commit/7e5a9c4d732aff56abfd7a31fce3b2f563d1f3f7))
* **test:** Add new population tests ([f4d618c](https://github.com/bxrne/darwin/commit/f4d618cbb603878a6a5754f56e2ea27bcd7f4028))
* **weigths:** Add crossover and mutate to weights Individual ([873453f](https://github.com/bxrne/darwin/commit/873453ff957d32fd91988d88bed22f0544b651cf))
* **wrapper:** Create action tree and weigths wrapper for init and running WIP ([b0bb9f7](https://github.com/bxrne/darwin/commit/b0bb9f7113d2503c6e4f94f1a1f6d161088bcaeb))


### Bug Fixes

* **darwin, config:** action tree ckpt ([2f37a71](https://github.com/bxrne/darwin/commit/2f37a71172c58efcef348d772796be2322e9326e))
* Got further ([1c4c40b](https://github.com/bxrne/darwin/commit/1c4c40b17881beb876539f9fac4c4885187891b9))
* **population:** GetPop no longer uses combined on return, CalcFit does not operate on nils now ([8993f9a](https://github.com/bxrne/darwin/commit/8993f9afe596ce760e9aa4456f9482f3b44ed732))
* **tests:** Make tests compliant with new weights contract ([3963cb5](https://github.com/bxrne/darwin/commit/3963cb564ee107d06a0e17de808f60a6c9f98264))
* **test:** Some tests fixed many more broken ([8db7ae9](https://github.com/bxrne/darwin/commit/8db7ae9578f4ff066bdbcad0bf206a84131e0c23))


### Code Refactoring

* **unused code:** Remove unused code ([6dac888](https://github.com/bxrne/darwin/commit/6dac8889779b4b3d694fe2380d7fd61634a0227b))

## [1.11.0](https://github.com/bxrne/darwin/compare/v1.10.0...v1.11.0) (2025-12-02)


### Features

* **action_tree:** Added cfg and tests ([d3e1147](https://github.com/bxrne/darwin/commit/d3e11471b0608b8082adb30c7d7ea6b0d070176b))
* **config, fitness:** Action tree softmax attempt ([0340e6f](https://github.com/bxrne/darwin/commit/0340e6fa42ebf543ffcf342f5f295c88705779c3))
* **fitness:** add action tree tcp client and receive obs for fitness ([de7496c](https://github.com/bxrne/darwin/commit/de7496c5bb0325d240fb743af0daab8727b6b17a))


### Bug Fixes

* **action_tree.go:** Add mutate and crossover ([1166336](https://github.com/bxrne/darwin/commit/1166336b18ec880fe2c4b55ca6f5e03d4cbe58ba))
* **cfg:** add tcp conn field ([55ab351](https://github.com/bxrne/darwin/commit/55ab35140708034e9d76db94641de7432423f2a9))
* **cmd, cfg:** Fix config test and call to RunEvo ([b9dd0f1](https://github.com/bxrne/darwin/commit/b9dd0f1c90f56e89282ec5fc77ccc34106f44e3f))
* **darwin, config:** action tree ckpt ([a3d4c45](https://github.com/bxrne/darwin/commit/a3d4c45ca939672c6fa3a600c655462afc9469de))
* **fitness:** cleanup and add better tests ([e8c98bf](https://github.com/bxrne/darwin/commit/e8c98bfc0bacf74dd05cb3e8efa26a4d3bf97a8c))
* **fitness:** remove switch for clamping ([08a5a5d](https://github.com/bxrne/darwin/commit/08a5a5d09324ed3f6702bf7c8542cf598470d357))
* Got further ([d2451c3](https://github.com/bxrne/darwin/commit/d2451c33db5faf3e053d78a6f7a9b055af8d3a06))
* **individual:** Fix create of ATIs to use vars as obs ([be2aff7](https://github.com/bxrne/darwin/commit/be2aff73a0ea080843ffbec4c227f050036703e3))


### Miscellaneous Chores

* **cmd, individual:** Add start of action tree with wm ([48d1b9a](https://github.com/bxrne/darwin/commit/48d1b9ac70b799f35d5a75fa7de51752bc1152b7))
* **cmd/demo:** Checkpointing a stubbed imp of ATI evolving ([2d41e29](https://github.com/bxrne/darwin/commit/2d41e295adc150f9acf5324e293087e489daf68e))
* **evolvable.go:** add package comment to silence gopls ([27371e0](https://github.com/bxrne/darwin/commit/27371e0095d75c95a0a1ed6bdcdf0e8ae54787c7))
* **fitness:** fix linting by checking client disconn err ([03e126e](https://github.com/bxrne/darwin/commit/03e126e0296c99e41982db15557893618f978da3))

## [1.10.0](https://github.com/bxrne/darwin/compare/v1.9.0...v1.10.0) (2025-11-21)


### Features

* **game:** functional bridge ([f50c3a3](https://github.com/bxrne/darwin/commit/f50c3a348c18174993223981e21544ca73fb97bd))


### Bug Fixes

* **game:** add numpy dep ([39b4eb2](https://github.com/bxrne/darwin/commit/39b4eb26cbceff6bd524fb98cec0c5537767ce58))


### Miscellaneous Chores

* **game:** add bash smoketest to bridge ([8157804](https://github.com/bxrne/darwin/commit/815780457a404d13e9269c466930212bfde34e4d))
* **game:** added generals usage example ([e0612da](https://github.com/bxrne/darwin/commit/e0612daaa32709f7fe1f677693abef1c929be4fc))
* **game:** initialise uv project ([f2fa3e4](https://github.com/bxrne/darwin/commit/f2fa3e4289c5dce90d1cefaa9fa7aac3d45d0f26))

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
