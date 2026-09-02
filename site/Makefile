.PHONY: serve build deps

# js.Build resolves the site's npm deps from the root node_modules
deps:
	$(MAKE) -C .. node_modules

serve: deps
	hugo server -p 3000

build: deps
	hugo --minify
