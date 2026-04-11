.PHONY: deploy

deploy:
	git pull --ff-only
	$(MAKE) -C broker env
	$(MAKE) -C broker up
	$(MAKE) -C auth env
	$(MAKE) -C auth migrate
	$(MAKE) -C auth up
	$(MAKE) -C message env
	$(MAKE) -C message migrate
	$(MAKE) -C message up
	$(MAKE) -C profile env
	$(MAKE) -C profile migrate
	$(MAKE) -C profile up
	$(MAKE) -C attachments env
	$(MAKE) -C attachments up
	$(MAKE) -C ui env
	$(MAKE) -C ui up
	$(MAKE) -C proxy env
	$(MAKE) -C proxy up
