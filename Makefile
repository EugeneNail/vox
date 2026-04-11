.PHONY: deploy

deploy:
	git pull --ff-only
	$(MAKE) -C broker networks
	$(MAKE) -C broker env
	$(MAKE) -C broker up
	$(MAKE) -C auth networks
	$(MAKE) -C auth env
	$(MAKE) -C auth db
	$(MAKE) -C auth migrate
	$(MAKE) -C auth up
	$(MAKE) -C message networks
	$(MAKE) -C message env
	$(MAKE) -C message db
	$(MAKE) -C message migrate
	$(MAKE) -C message up
	$(MAKE) -C profile networks
	$(MAKE) -C profile env
	$(MAKE) -C profile db
	$(MAKE) -C profile migrate
	$(MAKE) -C profile up
	$(MAKE) -C attachments networks
	$(MAKE) -C attachments env
	$(MAKE) -C attachments up
	$(MAKE) -C ui networks
	$(MAKE) -C ui env
	$(MAKE) -C ui up
	$(MAKE) -C proxy networks
	$(MAKE) -C proxy env
	$(MAKE) -C proxy up
