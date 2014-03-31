# Copyright (c) 2014 Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY: all clean nuke

all: editor scanner.go parser.go
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
	go clean
	rm -f *~ y.output y.go y.tab.c *.out ql.test

coerce.go: helper.go
	if [ -f coerce.go ] ; then rm coerce.go ; fi
	go run helper.go | gofmt > $@

cover:
	t=$(shell tempfile) ; go test -coverprofile $$t && go tool cover -html $$t && unlink $$t

editor: check scanner.go parser.go coerce.go
	go fmt
	go test -i
	go test
	go install

cpu: ql.test
	go test -c
	./$< -test.bench . -test.cpuprofile cpu.out
	go tool pprof $< cpu.out

mem: ql.test
	go test -c
	./$< -test.bench . -test.memprofile mem.out
	go tool pprof $< mem.out

nuke:
	go clean -i

parser.go: parser.y
	go tool yacc -o $@ -v /dev/null $<
	sed -i -e 's|//line.*||' -e 's/yyEofCode/yyEOFCode/' $@

ql.test: all

ql.y: doc.go
	sed -n '1,/^package/ s/^\/\/  //p' < $< \
		| ebnf2y -o $@ -oe $*.ebnf -start StatementList -pkg $* -p _

scanner.go: scanner.l parser.go
	golex -o $@ $<

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* *.go *.l parser.y || true
	@grep -n TODO *.go *.l parser.y testdata.ql || true
	@grep -n BUG *.go *.l parser.y || true
	@grep -n println *.go *.l parser.y || true

later:
	@grep -n LATER *.go *.l parser.y || true
	@grep -n MAYBE *.go *.l parser.y || true
