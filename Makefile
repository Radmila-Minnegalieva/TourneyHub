.DEFAULT_GOAL := help

.PHONY: help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-export docs-serve run

# Команды из корня выполняются в директории backend.
help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-export docs-serve run:
	$(MAKE) -C backend $@
