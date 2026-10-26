.DEFAULT_GOAL := help

.PHONY: help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-export docs-serve run integration-test metrics bootstrap

# Команды из корня выполняются в директории backend.
help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-export docs-serve run integration-test metrics bootstrap:
	$(MAKE) -C backend $@
