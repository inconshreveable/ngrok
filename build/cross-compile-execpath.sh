WORK=/var/folders/57/w68pgnks7txfk_r8g124dpmm0000gt/T/go-build594182218
mkdir -p $WORK/github.com/inconshreveable/

mkdir -p $WORK/github.com/inconshreveable/go-execpath/_obj/

pushd /Users/travis/build/inconshreveable/ngrok/src/github.com/inconshreveable/go-execpath

/usr/local/Cellar/go/1.1.1/pkg/tool/darwin_amd64/cgo -objdir $WORK/github.com/inconshreveable/go-execpath/_obj/ -- -I $WORK/github.com/inconshreveable/go-execpath/_obj/ darwin.go

/usr/local/Cellar/go/1.1.1/pkg/tool/darwin_amd64/8c -F -V -w -I $WORK/github.com/inconshreveable/go-execpath/_obj/ -I /usr/local/Cellar/go/1.1.1/pkg/darwin_386 -o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_defun.8 -D GOOS_darwin -D GOARCH_386 $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_defun.c

gcc -I . -g -O2 -fPIC -m32 -pthread -fno-common -print-libgcc-file-name

gcc -I . -g -O2 -fPIC -m32 -pthread -fno-common -I $WORK/github.com/inconshreveable/go-execpath/_obj/ -o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_main.o -c $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_main.c

gcc -I . -g -O2 -fPIC -m32 -pthread -fno-common -I $WORK/github.com/inconshreveable/go-execpath/_obj/ -o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_export.o -c $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_export.c

gcc -I . -g -O2 -fPIC -m32 -pthread -fno-common -I $WORK/github.com/inconshreveable/go-execpath/_obj/ -o $WORK/github.com/inconshreveable/go-execpath/_obj/darwin.cgo2.o -c $WORK/github.com/inconshreveable/go-execpath/_obj/darwin.cgo2.c

gcc -I . -g -O2 -fPIC -m32 -pthread -fno-common -o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_.o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_main.o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_export.o $WORK/github.com/inconshreveable/go-execpath/_obj/darwin.cgo2.o

/usr/local/Cellar/go/1.1.1/pkg/tool/darwin_amd64/cgo -objdir $WORK/github.com/inconshreveable/go-execpath/_obj/ -dynimport $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_.o -dynout $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_import.c

/usr/local/Cellar/go/1.1.1/pkg/tool/darwin_amd64/8c -F -V -w -I $WORK/github.com/inconshreveable/go-execpath/_obj/ -I /usr/local/Cellar/go/1.1.1/pkg/darwin_386 -o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_import.8 -D GOOS_darwin -D GOARCH_386 $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_import.c

gcc -I . -g -O2 -fPIC -m32 -pthread -fno-common -o $WORK/github.com/inconshreveable/go-execpath/_obj/_all.o $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_export.o $WORK/github.com/inconshreveable/go-execpath/_obj/darwin.cgo2.o -Wl,-r -nostdlib /usr/llvm-gcc-4.2/bin/../lib/gcc/i686-apple-darwin11/4.2.1/libgcc.a

/usr/local/Cellar/go/1.1.1/pkg/tool/darwin_amd64/8g -o $WORK/github.com/inconshreveable/go-execpath/_obj/_go_.8 -p github.com/inconshreveable/go-execpath -D _/Users/travis/build/inconshreveable/ngrok/src/github.com/inconshreveable/go-execpath -I $WORK ./execpath.go $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_gotypes.go $WORK/github.com/inconshreveable/go-execpath/_obj/darwin.cgo1.go

/usr/local/Cellar/go/1.1.1/pkg/tool/darwin_amd64/pack grcP $WORK $WORK/github.com/inconshreveable/go-execpath.a $WORK/github.com/inconshreveable/go-execpath/_obj/_go_.8 $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_import.8 $WORK/github.com/inconshreveable/go-execpath/_obj/_cgo_defun.8 $WORK/github.com/inconshreveable/go-execpath/_obj/_all.o

cp $WORK/github.com/inconshreveable/go-execpath.a /Users/travis/build/inconshreveable/ngrok/pkg/darwin_386/github.com/inconshreveable/go-execpath.a

popd
