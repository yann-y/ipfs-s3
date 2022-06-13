.PHONY: all github.com/yann-y/ipfs-s3 release-version

all: clean dist

version := $(shell cat VERSION)

#util/version.go:
release-version:
	git rev-parse HEAD|awk 'BEGIN {print "package util"} {print "const BuildGitVersion=\""$$0"\""} END{}' > util/version.go
	date +'%Y%m%d%H'| awk 'BEGIN{} {print "const BuildGitDate=\""$$0"\""} END{}' >> util/version.go

dist: github.com/yann-y/ipfs-s3
	mkdir -p build/github.com/yann-y/ipfs-s3-$(value version)
	cp script/run.sh build/github.com/yann-y/ipfs-s3-$(value version)/bin
	cp script/github.com/yann-y/ipfs-s3.cfg build/github.com/yann-y/ipfs-s3-$(value version)/bin
	cd build && tar cvzf github.com/yann-y/ipfs-s3.tar.gz github.com/yann-y/ipfs-s3-$(value version)
	#rm -r build/github.com/yann-y/ipfs-s3-$(value version)

github.com/yann-y/ipfs-s3: release-version
	mkdir -p build/github.com/yann-y/ipfs-s3-$(value version)/bin
	go build ${BUILD_FLAGS} -o build/github.com/yann-y/ipfs-s3-$(value version)/bin/github.com/yann-y/ipfs-s3

clean:
	rm -rf build

