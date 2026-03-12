.PHONY: console reset

console:
	sqlite3 ~/.config/watcher/watcher.db

DB := ~/.config/watcher/watcher.db

# make reset          - delete entire database
# make reset TABLE=x  - clear a specific table (threads, messages, users)
reset:
ifndef TABLE
	rm -f $(DB)
	@echo "database removed"
else
	sqlite3 $(DB) "DELETE FROM $(TABLE);"
	@echo "cleared table: $(TABLE)"
endif

pg:
	go run ./playground/main.go
