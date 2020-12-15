RUSTFMT_FLAGS = --config-path .rustfmt.toml --edition 2018
RUST_SOURCE_FILES = $(shell find . -name "*.rs" -not -path "./target/*")

.PHONY: rustfmt
rustfmt:
	rustfmt $(RUSTFMT_FLAGS) $(RUST_SOURCE_FILES)

.PHONY: check-rustfmt
check-rustfmt:
	rustfmt --check $(RUSTFMT_FLAGS) $(RUST_SOURCE_FILES)

.PHONY: test
test:
	cargo test
