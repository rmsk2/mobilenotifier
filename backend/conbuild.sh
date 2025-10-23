# Definitions
img=grap.de/notifier:v1
binary=notifier.tar
prog=notifier

# cleanup
rm $binary
rm $prog

swag init -g controller/swagger_base.go

# build binary
CGO_ENABLED=0 go build

# create image
myc=$(buildah from docker.io/library/alpine)
buildah copy $myc ./$prog /$prog
buildah config --entrypoint "/$prog" $myc
buildah config --port 5000 $myc
buildah commit $myc $img
buildah rm $myc
buildah push $img oci-archive:$binary:$img
buildah rmi $img

# Distribute image
scp $binary martin@debasus:$binary
scp $binary martin@debshuttle:$binary
scp $binary martin@debasus2:$binary

