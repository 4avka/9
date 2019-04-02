# Anti-Parallelisation Proof of Work

![](https://git.parallelcoin.io/com/assets/raw/branch/master/complexpow.jpg)

In the graph above, you can see the directed acyclic graph that defines the hashing process. CN7 is Cryptonote 7 version 2 and Ar is Argon 2i. Each of the things that send an arrow to another node are multiple hashes that are processed in the next step.

As you can see, there is three primary pathways of execution and they converge with three going to CN7 and then back out to several other functions.

This execution path requires the different hashing functions to load their results into pipes to other processing stages, and in this case requires at least three different result sets to zip back and forth between stages at any given moment, meaning at least 2-3 times as much interconnect bandwidth utilisation than if the graph was a single path.

All methods for creating ASICs involves parallelising the processing as much as possible. This structure forces extensive use of interconnects, the Argon hash strains memory bus, CN7 strains the processor cache, and so each type of process requires definitely a separate processor and thus at least one channel to pass results back and forth between processors.

The speed of these interconnect buses is slowing down even more than processing speeds are going up (due to quantum limits). Faster interconnects are far more expensive devices than the hashing units themselves.

The idea is that the potential benefits of optimising the processing devices are countered by the cost of connecting them, and that cost is amplified by parallel paths inside a serial process. Possibly the graph could be made more complex, such as to put Equihash as a third node and several other fash hashes interspersed. 

The final design will not be settled until the new servers are out of beta testing. There is no way that even if Parallelcoin had a big market cap that they could ASICify the algorithm in 6 months, and with a low market cap, we should have a full year's grace.

By which time PoW will be deprecated to a rate limiter with a graph based logic clock consensus platform base ready to start implementing ledgers.