RUSTFMT_FLAGS = --config-path ../.rustfmt.toml --edition 2021
RUST_SOURCE_FILES = $(shell find . -name "*.rs" -not -path "./target/*")

.PHONY: rustfmt
rustfmt:
	rustfmt $(RUSTFMT_FLAGS) $(RUST_SOURCE_FILES)

.PHONY: check-rustfmt
check-rustfmt:
	rustfmt --check $(RUSTFMT_FLAGS) $(RUST_SOURCE_FILES)

.PHONY: test
test:
	@cargo test

.PHONY: build
build:
	@cargo build --release

.PHONY: build-debug
build-debug:
	@cargo build

