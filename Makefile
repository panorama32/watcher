.PHONY: console reset

console:
	sqlite3 ~/.config/watcher/watcher.db

reset:
	rm -f ~/.config/watcher/watcher.db
	@echo "database removed"

pg:
	go run ./playground/main.go
