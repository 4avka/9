## <img src="https://git.parallelcoin.io/com/assets/raw/branch/master/logo/logo64x64.png"> <sup><i>9</i></sup> 


Parallelcoin all-in-one next generation server application suite.

The old version, [pod](https://git.parallelcoin.io/pod) has grown quite unruly and disorderly in an attempt to cleanly join together the three base btcsuite apps, the node, wallet and cli controller.

The catalyst was the eventual discovery that there is not one CLI library in existence that satisfactorily covers the need for configuration files, relatively easy reconfiguration from CLI, and the many generally overly complex ways these are implemented.

## The user does not want choices

Many conventions in software development stem from a time where IT was a far less vertically connected, tangled in the red tape of copyright, or sometimes, even, from bad analogies.

Command line interfaces originated with typewriters that triggered hammers up to hundreds of miles away, using what is called 'Teletype'.

For reasons of efficiency and unambiguity, certain conventions of message construction developed. Similar inter-operation systems can be found all over the place, resistor colour codes, NATO letter-words.

For a very long time, computer programmers were stuck at nasty little text console terminals, and as we see with Unix, the great majority of command line stuff is two letter commands and one letter flags, sometimes possible to collapse them to `-abcwtfbbq`.

So people sorta have got used to the idea of command line parameters having a certain shape. Even when there isn't really any real reason for it to continue.

So, the hallmark of generation ***9*** is going to be about radical rethinking, and especially, removal or embedding of complicated configuration junk that is rarely needed at the touch of some keys.

### The Feng Shui of Software

How we use computers now compared to in the past, and in different tech cultures, has changed a lot. So, the interface will be consciously designed to reflect actual use patterns.

1. The unadorned executable will launch the GUI, as that is most likely how it is wanted to be used. The GUI will have its own configuration, which will also reflect this philosophy of minimalism - and complexities will not be locked off but just placed deeper.
2. There will be a type-ahead partial match driven configuration interface that runs in a terminal and uses maybe 20 lines to render more useful information to the user. This will be the main way to change settings.
3. There is a simple launcher for node, wallet and combined shell, with no flags available except to change the profile directory path.
4. The CLI controller already has its own parser, so it will be used as is.
5. Aside from the main apps, there is also trigger functions that drop indexes, reset configuration, copy and generate testnet configurations.
6. All certificate and key generating steps will be automated as much as possible. TLS will be by default and the node will inform you about this.

