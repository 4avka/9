 <img src="https://git.parallelcoin.io/com/assets/raw/branch/master/logo/logo64x64.png"> <sup><i>9</i></sup><sub>[all in one everything]</sub>

# what is parallelcoin?

Parallelcoin is an altcoin that appeared around the time of the MtGox hack and the first appearance of ASIC miners on the market.

The original creator of the token shortly afterwards disappeared again, and hasn't been seen since.

It was adopted and after some time finally programmers were found to bring this old miner-focused coin into the next generation.

## roadmap

##### where we've been

1. [Original creation of coin by Parallaxis](https://bitcointalk.org/index.php?topic=721170.msg8145710#msg8145710), release 8 February 2014
2. Community Takeover Announcement https://bitcointalk.org/index.php?topic=1097017 23 June 2015
3. Loki Verloren [formally started working on a new server](https://bitcointalk.org/index.php?topic=1097017.msg39670320#msg39670320) 8 June 2018 based on btcsuite's [btcd](https://github.com/btcsuite/btcd)

### plan 9 from crypto space!

##### where we're going

We are very proud of the work so far and we think that people will love the new Parallelcoin:

- 9 complex multi-algorithm orthogonal complexity proof of work hash functions
    - cryptonote 7 v2 for cache-heavy hashing
    - aurora2i, memory bus intensive
    - + 9 different hash algorithms already with ASIC hardware
- 4 part averaging algorithm:
    - All time average - to prevent long term drift
    - One day trailing simple average
    - Per-block exponential weighted average
    - Per-algorithm equal spacing averaging with exponential weighted average
- strong resistance to timing and rhythm attacks by the use of multiple competing averagers 
    - no hard limiters
    - tends towards equilibrium
- default odds-following weighted randomising work schedulers
    - miners bias timing to the inner 2SD (9-27 seconds)
- 9 second blocks with difficulty reduction damper to reduce coincidence of very short intervals between blocks
    - increases warmup cost of specialised miners with a shorter average time between blocks
    - decreases window of opportunity for pool miners via their inherent high latency, and the same for botnets
    - short time between blocks tends to favour miners nearer to the origin of  the transaction - instead of going to the cheapest surplus electricity
    - more practical for face to face transactions due to shorter clearance time, without any additional systems overlaying it

1. Release of Plan 9 from Crypto Space software suite, mid April 2019
2. Hard Fork scheduled for ~ block 199999 ~ May 2019
3. 3 months intensive monitoring and bug-catching until end of July
4. Begin intensive work towards implementing the Distributed Journal Cache Protocol with its Proof of Causality logical clocks and probabalistic graph analysis, first working beta by January 2020
5. Migrate DUO ledger to DJCP and and start building out SDK and first applications.

## Application Feng Shui

Technologies develop slowly over time and they leave behind trails of arbitrary things that stuck despite no sensible reason for it. The carpal-tunnel-inducing intentionally difficult QWERTY keyboard, the kilogram, which still lacks a reference free derivation and upholds several pillars of physics, the heels on shoes that nobody uses to hang onto stirrups anymore...

Thus, every aspect we can examine and improve especially in user experience is under the microscope.

Take command line applications, for example: These were originally devised to be used on dumb terminals connected by slow analogue modem connections, to gigantic room and building sized computers with hefty megabytes of storage.

Even still, nowhere near enough people touch type, but it gets better with time. We are mostly used to even tapping away with two fingers at touch screens at a pace that gets within comfortable distance of hunt and peck typists.

The point is, all these things were influenced by conditions that no longer exist. We don't have the problem of colliding hammers on the typewriter anymore, instead we have other problems, like a keyboard layout designed to make typing slower and harder work.

### user types

#### casual users

These people will just download the binary and run it in their GUI environment. So when ***9*** is run without parameters, as happens from a double click, it launches the GUI, which will select the default profile directory for one user. For devs, this can also be launched with the path of another profile folder, for allowing use of testnets and so forth.

#### miner users

Miners will just want something simple and automatic that just works. For miners, running 1-3 full nodes to punch out their blocks, in a 50-200sqm space, all connected by 100/1000mbit ethernet, and located near a major optic fibre backbone. For them, the informative and mostly keyboard driven CLI interface for configuration will be a pleasure, and saves them wasting their time just to learn yet another arbitrarily complex and specific configuration scheme.

They are used to using ssh connections, and the simplicity of the miner work dispatch push subscriptions, and the confidence in knowing that what costs they lay down to mine this coin, will not be worthless overnight. To not have to read endless help files and search forums, it just works, it tells you everything you need it to and doesn't ask for rubbish you don't need.

Plus, because we wrote it in Go, there's more machines that can run it (easily) than non-Go crypto software, and you will be able to get it in binary form from day zero.


#### developers

These people are running exchanges, building websites, running websites, and so on. Though the configuration syntax is a tiny bit different from the usual, it is easy to understand, and the interactive CLI configuration doubles as a lightweight user manual. Such users may cry a bit about not being able to quickly change a setting right after watching a previous result.

But it's so simple to add such things and likely even some skulk around in the dark corners of the codebase.

But the nicest bit, for especially such as the authors of the software, is the testnet configuration tools. Print a score of default-based nodes all configured to only connect together, and then a second command starts it up and can rapidly set up arbitrary scenarios.

Plus, hopefully you'll enjoy that part so much you want to help make it better.

## Building

***9*** is built with Go 1.12 with modules fully enabled. You can just `go build` or `go install` in the root of the repository and voila.

## Documentation

See the [doc](doc/) directory for more information.