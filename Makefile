release:
	cd web && npm run build
	cd web && go-bindata-assetfs dist/
	mv web/bindata_assetfs.go .
	go build
	scp sirdm yixing@192.168.0.148:/usr/local/bin/

