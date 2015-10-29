
VERSION=$(shell git describe --tags --long)
DISTDIR=gvweb-$(VERSION)

.PHONY: gvweb dist gvweb-$(VERSION)
gvweb:
	go build --ldflags "-X main.g_Version=$(VERSION)" $@

gvweb-$(VERSION):
	mkdir -p $(DISTDIR)
	mkdir -p $(DISTDIR)/etc/
	cp gvweb $(DISTDIR)
	mkdir -p $(DISTDIR)/data/
	mkdir -p $(DISTDIR)/static/
	cp -ap etc/gvweb.service $(DISTDIR)/etc/
	cp -ap static/* $(DISTDIR)/static/
	tar --owner=nobody --group=nobody -cvzf $(DISTDIR).tar.gz $(DISTDIR)
	rm -rf $(DISTDIR)

dist: gvweb gvweb-$(VERSION)
