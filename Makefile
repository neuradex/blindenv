.PHONY: build test vet clean purge

build:
	go build -o ./blindenv .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f ./blindenv

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
