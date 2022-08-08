.PHONY: go-tidy
go-tidy: ## tidy all go modules
	$(call print-target)
	find . -name 'go.mod' -execdir go mod tidy -v \;

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


define print-target
	@printf "Executing target: \033[36m$@\033[0m\n"
endef