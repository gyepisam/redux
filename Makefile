# Yes, the irony is noted

.PHONY: clean all
exe = bin/redo bin/redo-ifchange bin/redo-init

all: $(exe)

bin/redo:
	go build -o $@ $(notdir $@)/main.go

bin/redo-ifchange:
	go build -o $@ $(notdir $@)/main.go

bin/redo-init:
	go build -o $@ $(notdir $@)/main.go


install: $(exe) 
	install -t /usr/local/bin/ $^

clean:
	#$(RM) *~ \#*\#
	find . -path ./t -prune -o -type f -name '*~' -exec rm -f {} \;
