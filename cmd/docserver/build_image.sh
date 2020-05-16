BUILDDIR="$(pwd)"
echo "BUILDDIR: $BUILDDIR"
echo "building docserver"
GOOS=linux go build -o docserver
echo "building mkdoc"
cd ../mkdoc && GOOS=linux go build -o mkdoc && mv mkdoc ../docserver
cd "$BUILDDIR"
echo "building docker image"
docker build -t thewinds/mkdoc-server .
echo "push docker image"
docker push thewinds/mkdoc-server
echo "clean up"
rm "$BUILDDIR/mkdoc"
rm "$BUILDDIR/docserver"