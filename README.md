Emoji-battle-royale
================

A golang webserver I'm using to host a battle royale style voteoff for emojis on my discord

Work in progress

    $ go run webserver.go
    
... and point a browser at http://localhost:8097
which returns an HTML webpage (home.html).

### Installation
I have tested the following installation on Ubuntu 18.04.1 x64 using a Digital Ocean dropplet

    $ cd /srv
    $ git clone https://github.com/cbahn/Emoji-battle-royale
    $ apt install --yes gcc
    $ snap install go --classic
    $ go get github.com/gorilla/mux
    $ cd Emoji-battle-royale
    $ go build webserver.go
    $ ./webserver

### Credits

Thanks to https://github.com/jimmahoney/golang-webserver for the awesome example server for me to start from