.PHONY: build test vet clean purge

build:
	go build -o ./blindenv .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f ./blindenv

# Remove all blindenv traces from the system (plugin cache, binary, PATH entries, config cache)
purge:
	@echo "=== Removing plugin cache ==="
	rm -rf $(HOME)/.claude/plugins/cache/blindenv
	@echo "=== Removing secret cache ==="
	rm -rf $(HOME)/.cache/blindenv
	@echo "=== Removing symlink ==="
	rm -f $(HOME)/.local/bin/blindenv
	@echo "=== Removing PATH entries from shell rc ==="
	@sed -i '' '/\[blindenv\] plugin bin/d' $(HOME)/.zshrc 2>/dev/null || true
	@sed -i '' '/\[blindenv\] plugin bin/d' $(HOME)/.bashrc 2>/dev/null || true
	@echo "=== Done. blindenv fully purged. ==="
