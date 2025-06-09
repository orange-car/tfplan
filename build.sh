#!/bin/sh
set -e

helpFunction()
{
   echo ""
   echo "Usage: $0 -r release"
   echo -e "\t-r Release version number"
   exit 1
}

test() {
  go test ./... -v --race --cover
}

build() {
  os=$1
  arch=$2
  release=$3
  
  GOOS=${os} GOARCH=${arch} go build -o bin/${release}/tfplan_${release}_${os}_${arch}
  tar -czvf bin/${release}/tfplan_${release}_${os}_${arch}.tar.gz bin/${release}/tfplan_${release}_${os}_${arch}
  rm bin/${release}/tfplan_${release}_${os}_${arch}
}

checksums() {
  release=$1
  rm -f ./bin/${release}/tfplan_${release}_checksums.txt
  touch ./bin/${release}/tfplan_${release}_checksums.txt

  for file in bin/${release}/*; do 
  echo "$file"
    if [[ $file == *.tar.gz ]]; then 
        md5sum ${file} >> bin/${release}/tfplan_${release}_checksums.txt
    fi 
done
}

getBranch() {
  CURRENT=$(git rev-parse --abbrev-ref HEAD)
}

checkout() {
  git checkout $1
}

execute() {
  echo "\033[1mcheckout out git tag ${RELEASE}...\033[0m"
  checkout $RELEASE
  echo ""
  echo "\033[1mtesting...\033[0m"
  echo ""
  test
  echo ""
  echo "\033[1mbuilding binaries for release ${RELEASE}...\033[0m"
  build darwin amd64 ${RELEASE}
  build darwin arm64 ${RELEASE}
  build linux amd64 ${RELEASE}
  build linux arm64 ${RELEASE}
  build windows amd64 ${RELEASE}
  build windows arm64 ${RELEASE}
  echo ""
  echo "\033[1mbuilding checksums...\033[0m"
  checksums ${RELEASE}
  checkout $CURRENT
}

while getopts "r:" opt
do
   case "$opt" in
      r ) RELEASE="$OPTARG" ;;
      ? ) helpFunction ;; # Print helpFunction in case parameter is non-existent
   esac
done

if [ -z "$RELEASE" ]
then
   echo "-r (release) cannot be empty";
   helpFunction
fi

execute