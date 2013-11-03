.PHONY: all clean nuke

all: editor scanner.go parser.go coerce.go
	go build
	go vet
	go install
	golint .
	make todo

bench: all
	go test -run NONE -bench .

check: ql.y
	go tool yacc -v /dev/null -o /dev/null $<

clean:
	@go clean
	rm -f *~ y.output y.go y.tab.c

coerce.go: helper.go
	if [ -f coerce.go ] ; then rm coerce.go ; fi
	go run helper.go -o $@

cover:
	t=$(shell tempfile) ; go test -coverprofile $$t && go tool cover -html $$t && unlink $$t

editor: check scanner.go parser.go
	go fmt
	go test -i
	go test

nuke:
	go clean -i

parser.go: parser.y
	go tool yacc -o $@ -v /dev/null $<
	sed -i -e 's|//line.*||' -e 's/yyEofCode/yyEOFCode/' $@

ql.y: doc.go
	sed -n '1,/^package/ s/^\/\/  //p' < $< \
		| ebnf2y -o $@ -oe $*.ebnf -start StatementList -pkg $* -p _

scanner.go: scanner.l parser.go
	golex -o $@ $<

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* *.go || true
	@grep -n TODO *.go || true
	@grep -n BUG *.go || true
	@grep -n println *.go || true

later:
	@grep -n LATER *.go || true
	@grep -n MAYBE *.go || true
