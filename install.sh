#!/usr/bin/env bash

GH_REPO="axllent/sndmail"
TIMEOUT=90

set -e

VERSION=$(curl --silent --location --max-time "${TIMEOUT}" "https://api.github.com/repos/${GH_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ $? -ne 0 ]; then
    echo -ne "\nThere was an error trying to check what is the latest version of sndmail.\nPlease try again later.\n"
    exit 1
fi

# detect the platform
OS="$(uname)"
case $OS in
Linux)
    OS='linux'
    ;;
FreeBSD)
    OS='freebsd'
    echo 'OS not supported'
    exit 2
    ;;
NetBSD)
    OS='netbsd'
    echo 'OS not supported'
    exit 2
    ;;
OpenBSD)
    OS='openbsd'
    echo 'OS not supported'
    exit 2
    ;;
Darwin)
    OS='darwin'
    ;;
SunOS)
    OS='solaris'
    echo 'OS not supported'
    exit 2
    ;;
*)
    echo 'OS not supported'
    exit 2
    ;;
esac

# detect the arch
OS_type="$(uname -m)"
case "$OS_type" in
x86_64 | amd64)
    OS_type='amd64'
    ;;
i?86 | x86)
    OS_type='386'
    ;;
aarch64 | arm64)
    OS_type='arm64'
    ;;
*)
    echo 'OS type not supported'
    exit 2
    ;;
esac

GH_REPO_BIN="sndmail-${OS}-${OS_type}.tar.gz"

#create tmp directory and move to it with macOS compatibility fallback
tmp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t 'sndmail-install.XXXXXXXXXX')
cd "$tmp_dir"

echo "Downloading sndmail $VERSION"
LINK="https://github.com/${GH_REPO}/releases/download/${VERSION}/${GH_REPO_BIN}"

curl --silent --location --max-time "${TIMEOUT}" "${LINK}" | tar zxf - || {
    echo "Error downloading"
    exit 2
}

cp sndmail /usr/local/sbin/ || exit 2
chmod 755 /usr/local/sbin/sndmail || exit 2
rm -f /usr/sbin/sendmail
ln -s /usr/local/sbin/sndmail /usr/sbin/sendmail

if [ ! -e "/etc/sndmail.conf" ]; then
    cp sndmail.conf.example /etc/sndmail.conf
fi

case "$OS" in
'linux')
    chown root:root /usr/sbin/sendmail || exit 2
    ;;
'freebsd' | 'openbsd' | 'netbsd' | 'darwin')
    chown root:wheel /usr/sbin/sendmail || exit 2
    ;;
*)
    echo 'OS not supported'
    exit 2
    ;;
esac

rm -rf "$tmp_dir"
echo "Linked /usr/sbin/sendmail to /usr/local/sbin/sndmail"
