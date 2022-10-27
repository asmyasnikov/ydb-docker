#!/usr/bin/env bash
set -ex

readonly WORKDIR="$(pwd)"
readonly OUTPUT_DIR="${WORKDIR}/.output"
readonly CACHE_DIR="${WORKDIR}/.cache"
readonly BUILD_DIR="${WORKDIR}/.build"
 
rm -rf \
 "${OUTPUT_DIR}" \
 "${CACHE_DIR}" \
 "${BUILD_DIR}" || :
mkdir -p "${CACHE_DIR}" "${BUILD_DIR}" "${OUTPUT_DIR}"

echo "##teamcity[buildStatus text='Building']"
sudo apt-get update
export DEBIAN_FRONTEND=noninteractive
sudo ln -sf /usr/share/zoneinfo/Etc/UTC /etc/localtime
sudo apt-get -y install tzdata
wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key|sudo apt-key add -
wget -O - https://apt.kitware.com/keys/kitware-archive-latest.asc |sudo apt-key add -
echo 'deb http://apt.kitware.com/ubuntu/ focal main' | sudo tee /etc/apt/sources.list.d/kitware.list >/dev/null
echo 'deb http://apt.llvm.org/focal/ llvm-toolchain-focal main' | sudo tee /etc/apt/sources.list.d/llvm.list >/dev/null

sudo apt-get update

sudo apt-get -y install git cmake python python3-pip ninja-build antlr3 m4 clang-12 lld-12 libidn11-dev libaio1 libaio-dev
sudo pip3 install conan

cd ${BUILD_DIR}
ls -lah ${WORKDIR}/
cmake -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_TOOLCHAIN_FILE=${WORKDIR}/clang.toolchain ${WORKDIR}
ninja

echo "##teamcity[buildStatus text='Archiving']"
readonly RELEASE_DIR="${WORKDIR}/.release/%release-name%"
readonly RELEASE_BIN_DIR="${RELEASE_DIR}/bin"
readonly RELEASE_LIB_DIR="${RELEASE_DIR}/lib"
mkdir -p "${RELEASE_DIR}"
if [[ -n "%bin-artifacts%" ]]; then
  mkdir -p "${RELEASE_BIN_DIR}"
fi
if [[ -n "%lib-artifacts%" ]]; then
  mkdir -p "${RELEASE_LIB_DIR}"
fi

for F in %bin-artifacts%; do
  if [[ -L "${F}" ]]; then
    F="$(readlink ${F})"
  fi
  DST="${RELEASE_BIN_DIR}/$(basename "${F}")"
  ln "${F}" "${DST}"
  strip "${DST}"
done

cd ..

for F in %artifacts%; do
  if [[ -L "${F}" ]]; then
    F="$(readlink ${F})"
  fi
  DST="${RELEASE_DIR}/$(basename "${F}")"
  ln "${F}" "${DST}"
done

for F in %lib-artifacts%; do
#  if [[ -L "${F}" ]]; then
#    F="$(readlink ${F})"
#  fi
  DST="${RELEASE_LIB_DIR}/$(basename "${F}")"
  cp -R "${F}" "${DST}"
#  strip "${DST}"
done

echo "##teamcity[setParameter name='result.release_name' value='%release-name%']"
echo "##teamcity[buildStatus text='%release-name%']"