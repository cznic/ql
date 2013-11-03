.PHONY: all clean nuke

all: editor
	go vet
	go install
	make todo

clean:
	go clean
	rm -f *~ ql

editor:
	go fmt
	go install

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* *.go || true
	@grep -n TODO *.go || true
	@grep -n BUG *.go || true
	@grep -n println *.go || true
