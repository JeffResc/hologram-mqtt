# Changelog

## [1.2.3](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.2.2...hologram-mqtt-v1.2.3) (2026-04-17)


### Miscellaneous

* **deps:** update golang:1.26-alpine docker digest to f853308 ([6b746d2](https://github.com/JeffResc/hologram-mqtt/commit/6b746d2d120fb5acadd5ef7e4421c14166c4c05b))
* **deps:** update golang:1.26-alpine docker digest to f853308 ([cb92f74](https://github.com/JeffResc/hologram-mqtt/commit/cb92f749e96facc2fb17a80ed76483817111fa0d))

## [1.2.2](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.2.1...hologram-mqtt-v1.2.2) (2026-04-16)


### Miscellaneous

* **deps:** update golang:1.26-alpine docker digest to 27f8293 ([fd68d0e](https://github.com/JeffResc/hologram-mqtt/commit/fd68d0eb6e90c136ee817184a7e304ad4b0ac9f8))
* **deps:** update golang:1.26-alpine docker digest to 27f8293 ([c185edb](https://github.com/JeffResc/hologram-mqtt/commit/c185edbb26a1ce20eb5c8c81cc0ca2e15b722448))
* **deps:** update googleapis/release-please-action digest to 5c625bf ([dd2756e](https://github.com/JeffResc/hologram-mqtt/commit/dd2756ecbd857502f69b697275326a208ca0b705))

## [1.2.1](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.2.0...hologram-mqtt-v1.2.1) (2026-04-13)


### Bug Fixes

* data race in paho MQTT client during integration tests ([850b1dd](https://github.com/JeffResc/hologram-mqtt/commit/850b1dd5e4f0b336526d74b28947ad667c1a5fbe))
* data race in paho MQTT client during integration tests ([370bcd6](https://github.com/JeffResc/hologram-mqtt/commit/370bcd6bd336fec8f6ed599fe9a9c5870bfd1648))

## [1.2.0](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.1.2...hologram-mqtt-v1.2.0) (2026-04-13)


### Features

* add .gitignore file ([d4fb884](https://github.com/JeffResc/hologram-mqtt/commit/d4fb884ff30d6774e66fdbf53791be648431c01d))
* add .gitignore file ([afa2f1f](https://github.com/JeffResc/hologram-mqtt/commit/afa2f1fefb4ef3d107b70528fec5c52c931cbb8e)), closes [#26](https://github.com/JeffResc/hologram-mqtt/issues/26)
* add command debouncing per device ([41c4616](https://github.com/JeffResc/hologram-mqtt/commit/41c4616d4cffc148128a3beb5b31f3135a5e6019))
* add command debouncing per device ([d6857b4](https://github.com/JeffResc/hologram-mqtt/commit/d6857b419125781ae9bceda33c694c9bfccc55ba)), closes [#57](https://github.com/JeffResc/hologram-mqtt/issues/57)
* add fuzz tests for external input parsing ([4c7d266](https://github.com/JeffResc/hologram-mqtt/commit/4c7d2668fc870ef186281c05af676a70d1d5765c))
* add fuzz tests for external input parsing ([87bc3ae](https://github.com/JeffResc/hologram-mqtt/commit/87bc3ae60c74703a577457270abd535384e48b6b)), closes [#34](https://github.com/JeffResc/hologram-mqtt/issues/34)
* add govulncheck vulnerability scanning to CI ([7bcb4df](https://github.com/JeffResc/hologram-mqtt/commit/7bcb4dff5543729e23675a707a9457296f3694ea))
* add govulncheck vulnerability scanning to CI ([b03c779](https://github.com/JeffResc/hologram-mqtt/commit/b03c7791c81c3d1f67e7f01baedede9e03558675)), closes [#29](https://github.com/JeffResc/hologram-mqtt/issues/29)
* add integration tests with a real MQTT broker ([50026fe](https://github.com/JeffResc/hologram-mqtt/commit/50026fef624221a0c48008b5318fc78473d7c2d0)), closes [#30](https://github.com/JeffResc/hologram-mqtt/issues/30)
* add integration tests with real MQTT broker ([fbe4473](https://github.com/JeffResc/hologram-mqtt/commit/fbe4473d3c3d15de8ffd1fc3e8644fffc49703f1))
* add optional PodDisruptionBudget to Helm chart ([c419ab5](https://github.com/JeffResc/hologram-mqtt/commit/c419ab5e96d3613cb05d524119e0cd6a4d187c88))
* add optional PodDisruptionBudget to Helm chart ([0283de2](https://github.com/JeffResc/hologram-mqtt/commit/0283de2e9a7491de66394f712de04ed75d3fe011))
* add Prometheus metrics ([1ff26a5](https://github.com/JeffResc/hologram-mqtt/commit/1ff26a53207b014ea7f8b16de49210e243a40d4f))
* add Prometheus metrics ([960d66d](https://github.com/JeffResc/hologram-mqtt/commit/960d66d54fa44111427e05fcd769b551019baee8)), closes [#58](https://github.com/JeffResc/hologram-mqtt/issues/58)
* add startup probe to Helm deployment ([e2f45da](https://github.com/JeffResc/hologram-mqtt/commit/e2f45dac80526b6a5139396a015acb56f4dc1c78))
* add startup probe to Helm deployment ([7581613](https://github.com/JeffResc/hologram-mqtt/commit/7581613b36427108cf8ed913550d0d3a9cbe4cea))
* enforce minimum test coverage threshold in CI ([f0605f5](https://github.com/JeffResc/hologram-mqtt/commit/f0605f5ba4a57c2e5e6c44bb448247daa96669d8))
* enforce minimum test coverage threshold in CI ([c655b51](https://github.com/JeffResc/hologram-mqtt/commit/c655b513ea396fd7ba16e3338a725725ff4be087)), closes [#35](https://github.com/JeffResc/hologram-mqtt/issues/35)
* run integration tests in CI ([9fae0d8](https://github.com/JeffResc/hologram-mqtt/commit/9fae0d8fc1c25569ff59f16f55da9d207dda1498))
* run integration tests in CI ([2c1a28c](https://github.com/JeffResc/hologram-mqtt/commit/2c1a28ca86a260fc34380951467297ab243e470f))
* set GOMAXPROCS automatically for containers ([ef77f4e](https://github.com/JeffResc/hologram-mqtt/commit/ef77f4ea7e4e61307c6bc25e1b5fdda3bbb8449f))
* set GOMAXPROCS automatically for containers ([1d54a1a](https://github.com/JeffResc/hologram-mqtt/commit/1d54a1a7a1c446419ec4b73c919b75409009cef9)), closes [#83](https://github.com/JeffResc/hologram-mqtt/issues/83)


### Bug Fixes

* add timeout to token.Wait() calls in MQTT onConnect ([b286cca](https://github.com/JeffResc/hologram-mqtt/commit/b286ccae4009a37632b3d7ef3d13aa0beaa8c50f))
* add timeout to token.Wait() calls in MQTT onConnect ([991a476](https://github.com/JeffResc/hologram-mqtt/commit/991a47652600ebd4774aa2ab697418480d5560c6))
* add timeouts to health check HTTP server ([4cfbb01](https://github.com/JeffResc/hologram-mqtt/commit/4cfbb01de62563b38256557c0cdf5c2db30043af))
* add timeouts to health check HTTP server ([d419202](https://github.com/JeffResc/hologram-mqtt/commit/d4192024c5a09a2508ef23f7df0ccc77a727ce86))
* check os.Unsetenv return and use t.Setenv in tests ([fdc0cba](https://github.com/JeffResc/hologram-mqtt/commit/fdc0cbae33960eb6002b6d145b320029c1f2a809))
* check resp.Body.Close error to satisfy errcheck linter ([a87a184](https://github.com/JeffResc/hologram-mqtt/commit/a87a184010eafcdf05eeeae078b7559e325d693b))
* close TOCTOU gap in command debounce logic ([39cca8d](https://github.com/JeffResc/hologram-mqtt/commit/39cca8d0db4c2f40fd4cd7797a2074c624c18c93))
* close TOCTOU gap in command debounce logic ([44e2257](https://github.com/JeffResc/hologram-mqtt/commit/44e2257bdfebb029bf97c397e4bfbf6d846c09a2))
* compute newCount from actual new devices instead of arithmetic ([bbea00e](https://github.com/JeffResc/hologram-mqtt/commit/bbea00e60041c8c1e8b14000af8016a1debb6d7a))
* correct negative newCount in poll log message ([48b3dae](https://github.com/JeffResc/hologram-mqtt/commit/48b3daee1a954e3a9cf9b84462eda922e9bb1bd7))
* data race in integration test format argument ([fd66a2f](https://github.com/JeffResc/hologram-mqtt/commit/fd66a2ff837825f3cf69a8205f1b0fb67519961a))
* data race in integration test format argument ([e4a8169](https://github.com/JeffResc/hologram-mqtt/commit/e4a8169ace96b0ef65ee3620193b74ca6f861ab9))
* default serviceAccount.automount to false in Helm chart ([8230c8f](https://github.com/JeffResc/hologram-mqtt/commit/8230c8f6ad0011f22cdf404642e9c9b06b9f8296))
* default serviceAccount.automount to false in Helm chart ([0b7323e](https://github.com/JeffResc/hologram-mqtt/commit/0b7323ec3f42013cf959dc2150db7eaab6f25b27))
* expand .dockerignore to reduce build context ([8859c95](https://github.com/JeffResc/hologram-mqtt/commit/8859c9508b6be48767a5b4941f1014c8de8e204a))
* expand .dockerignore to reduce build context ([007be9d](https://github.com/JeffResc/hologram-mqtt/commit/007be9d45a93670323b305ffbb6942ece905f12b)), closes [#74](https://github.com/JeffResc/hologram-mqtt/issues/74)
* expand golangci-lint config and fix findings ([e5c1f4e](https://github.com/JeffResc/hologram-mqtt/commit/e5c1f4e6c089d61f54b832314f8759c6336a7d13)), closes [#75](https://github.com/JeffResc/hologram-mqtt/issues/75)
* expand golangci-lint configuration and fix findings ([1af866b](https://github.com/JeffResc/hologram-mqtt/commit/1af866b3c88497dd5dd734ed769634b3e43e6b4f))
* gracefully drain in-flight commands on shutdown ([816fa9a](https://github.com/JeffResc/hologram-mqtt/commit/816fa9a3325d5f0057376cf11fa42e8b09702cd8))
* gracefully drain in-flight commands on shutdown ([c66c8ef](https://github.com/JeffResc/hologram-mqtt/commit/c66c8effcba9c6606c4c262189675909a7ef7f22)), closes [#55](https://github.com/JeffResc/hologram-mqtt/issues/55)
* improve health check to reflect poll health ([245aeea](https://github.com/JeffResc/hologram-mqtt/commit/245aeea0fb0eab113736f9931fcacdbfc0553678))
* improve health check to reflect poll health ([d832389](https://github.com/JeffResc/hologram-mqtt/commit/d832389b439e60f1a2317eb713abe04ba661ae69)), closes [#54](https://github.com/JeffResc/hologram-mqtt/issues/54)
* limit API response body reads to 10 MB ([adc8dff](https://github.com/JeffResc/hologram-mqtt/commit/adc8dffbd4f9e38a6f25d4545beb25de6f4d5084))
* limit response body size in Hologram API client ([952a7ee](https://github.com/JeffResc/hologram-mqtt/commit/952a7ee01ea6302ade53c6ecbdfddc5ead97a695))
* make retry backoffs configurable to speed up tests ([bcb3443](https://github.com/JeffResc/hologram-mqtt/commit/bcb344357b5db098a18f0a98f82c2f21dcaa9a0e))
* make retry backoffs configurable to speed up tests ([ffec3f8](https://github.com/JeffResc/hologram-mqtt/commit/ffec3f8da8e7c4d5329d52092fcc960362ca4f50)), closes [#59](https://github.com/JeffResc/hologram-mqtt/issues/59)
* pin GitHub Actions in vulnerability-scan CI job ([bfb9aea](https://github.com/JeffResc/hologram-mqtt/commit/bfb9aea9919569d0dbc75e643aba2b37d20b55c6))
* pin GitHub Actions to commit SHAs ([2e923ba](https://github.com/JeffResc/hologram-mqtt/commit/2e923baf0d2dccaa32002effddee4f65b34ba934))
* pin GitHub Actions to commit SHAs ([ae56d52](https://github.com/JeffResc/hologram-mqtt/commit/ae56d52f17aa10cf41c2d24def24ce5ae8c4c202)), closes [#27](https://github.com/JeffResc/hologram-mqtt/issues/27)
* pin GitHub Actions to SHA hashes in vulnerability-scan job ([95bb608](https://github.com/JeffResc/hologram-mqtt/commit/95bb608b395555ecd8cbee909e184f979d3d170c))
* pin Go builder image to digest in Dockerfile ([954eaf0](https://github.com/JeffResc/hologram-mqtt/commit/954eaf03872b7269102a117565279ace7fc4a4a7))
* pin Go builder image to digest in Dockerfile ([b9f9d32](https://github.com/JeffResc/hologram-mqtt/commit/b9f9d32ec1a122871683c25be8090b03ba450a6b)), closes [#28](https://github.com/JeffResc/hologram-mqtt/issues/28)
* pin golangci-lint to specific version in CI ([0941d3d](https://github.com/JeffResc/hologram-mqtt/commit/0941d3d626bc1a9dfbde82baf6a9fd2b552962d4))
* pin golangci-lint to v2.11.4 in CI ([8908fb5](https://github.com/JeffResc/hologram-mqtt/commit/8908fb55962fa6e368ff9a6a9fa1517bcfdbbbb7)), closes [#33](https://github.com/JeffResc/hologram-mqtt/issues/33)
* pin govulncheck version in CI ([37c9722](https://github.com/JeffResc/hologram-mqtt/commit/37c9722d508d6001293e4052cfc69b6d03cdf986))
* pin govulncheck version in CI ([491d8b1](https://github.com/JeffResc/hologram-mqtt/commit/491d8b1e083f4df00b1589428804dc0f5ac254bb))
* pin helm-unittest plugin version in CI ([57fa892](https://github.com/JeffResc/hologram-mqtt/commit/57fa8928f3a3e6ef20cd54c2ad8a292f1042a2a1))
* pin helm-unittest plugin version in CI ([9d0381a](https://github.com/JeffResc/hologram-mqtt/commit/9d0381addc424eee643a470c1273c72694d6dcdd))
* propagate context in MQTT command handler ([d3c7f9d](https://github.com/JeffResc/hologram-mqtt/commit/d3c7f9d3450aae97f2dd8ae9787d427348cbd403))
* propagate context in MQTT command handler ([dc1f3e6](https://github.com/JeffResc/hologram-mqtt/commit/dc1f3e69c6167bc368d0044ac354e359633c29a8)), closes [#31](https://github.com/JeffResc/hologram-mqtt/issues/31)
* remove broken healthcheck from docker-compose ([3b505b9](https://github.com/JeffResc/hologram-mqtt/commit/3b505b92b5096211f2af153b406c7faf32b2cc8c))
* remove broken healthcheck from docker-compose ([5c645fc](https://github.com/JeffResc/hologram-mqtt/commit/5c645fcc04dc5c07ca191d56b32af2bf9e225a8c)), closes [#78](https://github.com/JeffResc/hologram-mqtt/issues/78)
* remove raw API response debug log ([788f5a0](https://github.com/JeffResc/hologram-mqtt/commit/788f5a064b1bbf7939b59c4d5091e17482b4a084))
* remove stored context from Bridge struct ([49d2c29](https://github.com/JeffResc/hologram-mqtt/commit/49d2c296b03c9ae4a8aa91eb1f7da2fe6d657b59))
* remove stored context from Bridge struct ([13008a4](https://github.com/JeffResc/hologram-mqtt/commit/13008a4541f7e726d8d8deb246b887aacebea270))
* replace raw API response debug log with byte count ([b3b49b5](https://github.com/JeffResc/hologram-mqtt/commit/b3b49b56925dea43898e3a8eb90f750dff68bd72))
* restore --verify=false for helm-unittest plugin install ([8505ade](https://github.com/JeffResc/hologram-mqtt/commit/8505ade57f33c8ed2f8c66e2ae69cf1b8a3af38e))
* retry transient errors in Hologram API client ([df7904a](https://github.com/JeffResc/hologram-mqtt/commit/df7904a93bfafb91ef202417e8e826abafa0ddc4))
* retry transient errors in Hologram API client ([cb9b56b](https://github.com/JeffResc/hologram-mqtt/commit/cb9b56bc67cc8ee5baad535e768eb2a1d2771e2a))
* return error for invalid HOLOGRAM_ORG_ID ([1b8ee26](https://github.com/JeffResc/hologram-mqtt/commit/1b8ee26fd56247359c765fac7a3538d005735d4d))
* return error for invalid HOLOGRAM_ORG_ID ([12c5215](https://github.com/JeffResc/hologram-mqtt/commit/12c5215f4077449d4f73e57b65ec3084d2382430)), closes [#32](https://github.com/JeffResc/hologram-mqtt/issues/32)
* return error for invalid POLL_INTERVAL ([e7c842b](https://github.com/JeffResc/hologram-mqtt/commit/e7c842b4b9284fa6c6b3c114c7452c5a738ce6ed))
* return error for invalid POLL_INTERVAL ([7928181](https://github.com/JeffResc/hologram-mqtt/commit/7928181c774e5de5945e6b738f73c20568199c3a)), closes [#53](https://github.com/JeffResc/hologram-mqtt/issues/53)
* scope .gitignore binary entry to repo root ([2171af3](https://github.com/JeffResc/hologram-mqtt/commit/2171af34cb774e9d0c0cb809fe2f44228c49a5b7))
* scope .gitignore binary entry to repo root ([913f6bd](https://github.com/JeffResc/hologram-mqtt/commit/913f6bda2fcaa604f5a6351b7a1e5d76a2e57eff)), closes [#72](https://github.com/JeffResc/hologram-mqtt/issues/72)
* surface MQTT re-subscription failures in health check ([08351eb](https://github.com/JeffResc/hologram-mqtt/commit/08351eb07e3386d86742060ccf229226cb037416))
* surface MQTT re-subscription failures in health check ([8fedc6b](https://github.com/JeffResc/hologram-mqtt/commit/8fedc6b34087aa5fe52fe4e5bf2fc1e4545f1411)), closes [#56](https://github.com/JeffResc/hologram-mqtt/issues/56)
* use alternating commands in concurrent test ([e525b14](https://github.com/JeffResc/hologram-mqtt/commit/e525b146710fa2d948e0611b0695bb8146f3cc91))
* use alternating commands in concurrent test ([02ebf46](https://github.com/JeffResc/hologram-mqtt/commit/02ebf46c0777ed082901404a2bde3768fec7c86e)), closes [#79](https://github.com/JeffResc/hologram-mqtt/issues/79)
* use graceful shutdown for health HTTP server ([7463a07](https://github.com/JeffResc/hologram-mqtt/commit/7463a0757773be14e4ab12d7ae59be9aee38ede9))
* use graceful shutdown for health HTTP server ([d806c58](https://github.com/JeffResc/hologram-mqtt/commit/d806c5837d12823dabe8488af91b7910828aa0bf)), closes [#73](https://github.com/JeffResc/hologram-mqtt/issues/73)
* validate MQTT topic prefix is non-empty ([acc758c](https://github.com/JeffResc/hologram-mqtt/commit/acc758ccfc6dd2416199c36e951168ad485b603d))
* validate MQTT topic prefix is non-empty ([8681b30](https://github.com/JeffResc/hologram-mqtt/commit/8681b30ac743768b76634d2cb057bc72791226ce)), closes [#60](https://github.com/JeffResc/hologram-mqtt/issues/60)


### Miscellaneous

* add PR and issue templates ([b1e8c5b](https://github.com/JeffResc/hologram-mqtt/commit/b1e8c5b99946e354217c7489d05ce77c1cdeaf91))
* add PR and issue templates ([eae7e82](https://github.com/JeffResc/hologram-mqtt/commit/eae7e82124b2c5bd91d869abf49fde3ce477a9a0))
* switch license from MIT to Apache 2.0 ([b49e498](https://github.com/JeffResc/hologram-mqtt/commit/b49e4982192fa946373ed570e378f00c3eb5c831))
* switch license from MIT to Apache 2.0 ([0902f26](https://github.com/JeffResc/hologram-mqtt/commit/0902f26162e81dbb3767696271d0460a40fd0113))


### Documentation

* add CONTRIBUTING.md ([cc8d12f](https://github.com/JeffResc/hologram-mqtt/commit/cc8d12f3bffe055a8f496f7375e338e63b69597d))
* add CONTRIBUTING.md ([00d8013](https://github.com/JeffResc/hologram-mqtt/commit/00d801398fc87b13669b179e549618e367368ed2)), closes [#38](https://github.com/JeffResc/hologram-mqtt/issues/38)
* add metrics section to README ([3c8dab2](https://github.com/JeffResc/hologram-mqtt/commit/3c8dab2a2d7221c828a56d06e048cca42d88dd9a))
* add metrics section to README ([42c8c83](https://github.com/JeffResc/hologram-mqtt/commit/42c8c832514a5c7bdab949107157952f76fe33cc)), closes [#77](https://github.com/JeffResc/hologram-mqtt/issues/77)
* add Mosquitto config for docker-compose ([ecbac11](https://github.com/JeffResc/hologram-mqtt/commit/ecbac11217d5050c74918545d37dd3c30ea05a3b))
* add Mosquitto config for docker-compose ([5e07847](https://github.com/JeffResc/hologram-mqtt/commit/5e07847813c74221623353138d0b7186d1be8c4e))
* add SECURITY.md ([ce8a30c](https://github.com/JeffResc/hologram-mqtt/commit/ce8a30c401a539689a8ffdc323561fb5f65fe228))
* add SECURITY.md ([2fd6111](https://github.com/JeffResc/hologram-mqtt/commit/2fd61113a0c2d219d55ad66f710ce941f4d953d0)), closes [#37](https://github.com/JeffResc/hologram-mqtt/issues/37)
* document validation constraints in config.example.yaml ([2fa1a4d](https://github.com/JeffResc/hologram-mqtt/commit/2fa1a4de81166c1caec3b755c954fe90d57ccc66))
* document validation constraints in config.example.yaml ([c5ddec5](https://github.com/JeffResc/hologram-mqtt/commit/c5ddec5ac17e310b03db978d162b902266893405)), closes [#76](https://github.com/JeffResc/hologram-mqtt/issues/76)

## [1.1.2](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.1.1...hologram-mqtt-v1.1.2) (2026-04-10)


### Miscellaneous

* **deps:** update dependency go to v1.26.2 ([44a6cf9](https://github.com/JeffResc/hologram-mqtt/commit/44a6cf9e5c18ef9caa873e3b402cdd8654729ebe))

## [1.1.1](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.1.0...hologram-mqtt-v1.1.1) (2026-04-03)


### Bug Fixes

* add org ID configuration to Helm chart ([a7c3a10](https://github.com/JeffResc/hologram-mqtt/commit/a7c3a100ed8d7e9120110767bfdd5e20791c49c8))

## [1.1.0](https://github.com/JeffResc/hologram-mqtt/compare/hologram-mqtt-v1.0.0...hologram-mqtt-v1.1.0) (2026-04-03)


### Features

* add Helm chart OCI release to GitHub Actions and Kubernetes docs ([1df7e63](https://github.com/JeffResc/hologram-mqtt/commit/1df7e63c8138ecc71f3efd344bdce9820d4b50ac))


### Bug Fixes

* lowercase repository owner in Helm OCI push ([d053b67](https://github.com/JeffResc/hologram-mqtt/commit/d053b67c21ea890ca9d4fb3b3a76b4b71607d01e))
* use raw version tags for Docker image to ensure proper OCI tagging ([9d05678](https://github.com/JeffResc/hologram-mqtt/commit/9d05678b144dc0910fd858f7334c34aff1648232))


### Miscellaneous

* **deps:** update actions/upload-artifact action to v7 ([d96820b](https://github.com/JeffResc/hologram-mqtt/commit/d96820b5ddf3edb1f83d743b208be5ed90e09368))
* **deps:** update azure/setup-helm action to v5 ([982c4b6](https://github.com/JeffResc/hologram-mqtt/commit/982c4b6a2078f1d50d7c9cbabdaa2a18e702b220))
* **deps:** update docker/build-push-action action to v7 ([e53c953](https://github.com/JeffResc/hologram-mqtt/commit/e53c9539e8b7f5da47bfef6e67a29289c855360e))
* **deps:** update docker/login-action action to v4 ([60f2a11](https://github.com/JeffResc/hologram-mqtt/commit/60f2a11511d012bc803481ef38651993168a2a97))
* **deps:** update docker/metadata-action action to v6 ([5a13a0a](https://github.com/JeffResc/hologram-mqtt/commit/5a13a0ae6ef8dcc4e8234dc578b72afc6c07227d))
* **deps:** update docker/setup-buildx-action action to v4 ([90d4ef6](https://github.com/JeffResc/hologram-mqtt/commit/90d4ef659a3379615be01c10b6cb999d8dcc32b8))
* **deps:** update docker/setup-qemu-action action to v4 ([562dadd](https://github.com/JeffResc/hologram-mqtt/commit/562dadd1f9c45c0c344ea39dfa9cde62322d97cb))
* **deps:** update golangci/golangci-lint-action action to v9 ([72b89fe](https://github.com/JeffResc/hologram-mqtt/commit/72b89fe403479c8e6d2b7e445f5eb462c252a6a7))
* **main:** release hologram-mqtt 1.1.0 ([6cfa966](https://github.com/JeffResc/hologram-mqtt/commit/6cfa9666651b3e408aca8d135565c59835e27e68))

## 1.0.0 (2026-04-03)


### Features

* add concurrency safety, health check, TLS support, and edge case tests ([1f253f0](https://github.com/JeffResc/hologram-mqtt/commit/1f253f05a60110b12d27fdb9aa0210d480a0e3cb))
* add Docker Compose template and Helm chart ([3ba330a](https://github.com/JeffResc/hologram-mqtt/commit/3ba330ace97c73459c01bdd1d7e6e931fc4d9c0c))
* add HOLOGRAM_ORG_ID config for multi-org API keys ([65aae19](https://github.com/JeffResc/hologram-mqtt/commit/65aae195b458b2abee48979b772f3e74c6c6528a))
* add multi-arch binary and Docker image builds ([ca1373c](https://github.com/JeffResc/hologram-mqtt/commit/ca1373c9ed75a91af329a80ae4fdc3b101e23e3b))
* align types with real Hologram API response, add metadata ([ee8ebbe](https://github.com/JeffResc/hologram-mqtt/commit/ee8ebbe465764d15052a2e28413867cd5d48d400))
* implement Hologram.io MQTT bridge for Home Assistant ([cbb1ce6](https://github.com/JeffResc/hologram-mqtt/commit/cbb1ce65b8fd9042bfaafa04bd55d95e75d17cf6))
* implement Hologram.io MQTT bridge for Home Assistant ([4d662c4](https://github.com/JeffResc/hologram-mqtt/commit/4d662c4f7f2d62c314136c4bbd61e63a32ed42c7))


### Bug Fixes

* change default MQTT topic prefix from hologram-mqtt to hologram ([4fb871d](https://github.com/JeffResc/hologram-mqtt/commit/4fb871d27989a602e094e2adc17f2cdfab695921))
* check error returns from w.Write in health handler ([b17d1fc](https://github.com/JeffResc/hologram-mqtt/commit/b17d1fc78cf0b4be20801ccb65f83dbde16f1396))
* handle errcheck lint errors for Close() calls ([47bb491](https://github.com/JeffResc/hologram-mqtt/commit/47bb491b45732f12f7a7fd6669fec81840957f5f))
* populate attributes from cellular link when device fields are empty ([9f92f36](https://github.com/JeffResc/hologram-mqtt/commit/9f92f36f6118a3a8c792590afd45d61286641a7a))
* resolve golangci-lint errors ([42b3528](https://github.com/JeffResc/hologram-mqtt/commit/42b3528b86e0e75acfec9a84f4fc7c578379f3ed))
* unmarshal links.cellular as array of objects ([a31bd17](https://github.com/JeffResc/hologram-mqtt/commit/a31bd17a9c65994db938866f5804845e5216c944))
* use correct types for IMSI and carrier fields ([7654eca](https://github.com/JeffResc/hologram-mqtt/commit/7654eca8d2d6a09ae2681d447f70913384ba1e18))
* use minor version bumps for pre-1.0.0 releases ([0df48e9](https://github.com/JeffResc/hologram-mqtt/commit/0df48e91776156002005a61ed4122a96173726f8))
* use pinned distroless static image digest ([be58e9d](https://github.com/JeffResc/hologram-mqtt/commit/be58e9d680200d608344001c01952bdfb7320e06))


### Miscellaneous

* **deps:** update actions/checkout action to v6 ([2758e92](https://github.com/JeffResc/hologram-mqtt/commit/2758e92794b70c9e1065f7716aaf5cdf9c32f3d2))
* **deps:** update actions/checkout action to v6 ([4c0d11d](https://github.com/JeffResc/hologram-mqtt/commit/4c0d11d25fc4eacc9b9bcf2078f229e261168d04))
* **deps:** update actions/setup-go action to v6 ([bcec6a0](https://github.com/JeffResc/hologram-mqtt/commit/bcec6a007717e474d611f6c0d4499194c61e6244))
* **deps:** update actions/setup-go action to v6 ([a798b1a](https://github.com/JeffResc/hologram-mqtt/commit/a798b1a0a120161069911c2f9fcf4ede947bf3ca))
* **deps:** update dependency go to 1.26 ([62dd92b](https://github.com/JeffResc/hologram-mqtt/commit/62dd92be97a01a9af2685b87a02209e3bf8b82c5))
* **deps:** update go.mod to go 1.26 and combine renovate branches ([b43456e](https://github.com/JeffResc/hologram-mqtt/commit/b43456e5bae4a0f0aaae48c3c91b4298bc5cbfa5))
* **deps:** update golang docker tag to v1.26 ([b88834a](https://github.com/JeffResc/hologram-mqtt/commit/b88834a9178d3d4ea617afb96f26ac7a2336b1b7))
* **deps:** upgrade golangci-lint to v2 for Go 1.26 support ([47c3cf6](https://github.com/JeffResc/hologram-mqtt/commit/47c3cf6aca27c5ef85c69178e4242a539fe20c67))
