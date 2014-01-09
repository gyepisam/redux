# Yes, the irony is noted

.PHONY: clean all

all: bin/redo bin/redo-ifchange

bin/redo:
	go build -o $@ $(notdir $@)/main.go

bin/redo-ifchange:
	go build -o $@ $(notdir $@)/main.go

install: bin/redo bin/redo-ifchange
	install -t /usr/local/bin/ $^

clean:
	#$(RM) *~ \#*\#
	find . -path ./t -prune -o -type f -name '*~' -exec rm -f {} \;
