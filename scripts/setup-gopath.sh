set -x

mkdir -p ~/bin ~/cache
export PATH=~/bin:$PATH

export GOSPACE_ROOT=$GOPATH
export GOSPACE_PKG=czarcoin.org/czarcoin
export GOSPACE_REPO=git@github.com:czarcoin/czarcoin/git

# setup gospace
wget -O ~/bin/gospace https://github.com/czarcoin/gospace/releases/download/v0.0.5/gospace_linux_amd64
chmod +x ~/bin/gospace

# find module dependency hash
MODHASH=$(gospace hash)

# download dependencies, if we don't have them in cache
if [ ! -f $HOME/cache/$MODHASH.zip ]; then
    gospace zip-vendor $HOME/cache/$MODHASH.zip
fi

# unpack the dependencies into gopath
gospace unzip-vendor $HOME/cache/$MODHASH.zip
gospace flatten-vendor

set +x
