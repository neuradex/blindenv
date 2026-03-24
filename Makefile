VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)

.PHONY: build test vet clean purge bump

build:
	go build -ldflags "-X github.com/neuradex/blindenv/cmd.version=$(VERSION)" -o ./blindenv .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f ./blindenv

# Usage: make bump v=0.4.0
bump:
ifndef v
	$(error Usage: make bump v=0.4.0)
endif
	@sed -i '' 's/"version": "[^"]*"/"version": "$(v)"/' .claude-plugin/plugin.json .claude-plugin/marketplace.json
	@echo "Bumped to $(v)"
	@echo "  git add -A && git commit -m 'chore: v$(v)' && git tag v$(v) && git push origin main --tags"

# Reset blindenv install state for testing (run /plugin uninstall blindenv@blindenv first)
purge:
	@echo "⚠  Run '/plugin uninstall blindenv@blindenv' in Claude Code first!"
	@echo ""
	@echo "=== Removing secret cache ==="
	rm -rf $(HOME)/.cache/blindenv
	@echo "=== Removing symlink ==="
	rm -f $(HOME)/.local/bin/blindenv
	@echo "=== Removing PATH entries from shell rc ==="
	@sed -i '' '/\[blindenv\] plugin bin/d' $(HOME)/.zshrc 2>/dev/null || true
	@sed -i '' '/\[blindenv\] plugin bin/d' $(HOME)/.bashrc 2>/dev/null || true
	@echo ""
	@echo "=== Done. Now reinstall with: ==="
	@echo "  /plugin install blindenv@blindenv"
