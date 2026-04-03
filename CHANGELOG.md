# Changelog

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
