# Makefile

.PHONY: test

define ensure-gotest
	@if [ ! -f gotest ]; then \
	    echo "Downloading gotest..."; \
	    curl  https://gotest-release.s3.amazonaws.com/gotest_linux > gotest && chmod +x gotest; \
	fi
endef

test: ensure-gotest
	@bash scripts/run-test-local

ensure-gotest:
	$(ensure-gotest)
clean:
	rm -f ./gotest
