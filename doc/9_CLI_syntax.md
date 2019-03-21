# <img src="https://git.parallelcoin.io/com/assets/raw/branch/master/logo/logo64x64.png"> <sup><i>9</i></sup> 
> **Parallelcoin** CLI syntax documentation

### All the things it can do

launch gui (default datadir)

    9

launch gui (other datadir)

    9 (datadir)

show help

    9 (datadir) [help|h]
    9 [help|h]

launch configuration cli

    9 (datadir) [conf]
    9 [conf]

load rpc cli

    9 (datadir) [cli]
    9 [cli]

show listcommands

    9 (datadir) (listcommands|list|l)
    9 [listcommands|list|l]

query from full node

	9 (datadir) [rpc|r] (getinfo/blahblah...)
	9 [rpc|r] (getinfo/blahblah...)

run full node

	9 (datadir) [node|n]
	9 [node|n]

run wallet

	9 (datadir) [node|n]
	9 [node|n]

run shell

	9 (datadir) [shell|s]
	9 [shell|s]

reset to factory defaults

	9 (datadir) [reinit]
	9 [reinit]

new datadir at ./test

	9 [new]

new datadir default at ./test

	9 [new] (number)

number new datadir with basename

	9 [new] (basename) (number)
		
1 new datadir with basename		

	9 [new] (basename)

copy from home to datadir at ./test

	9 (datadir) [copy]
	9 [copy]

copy datadir at ./test

	9 (datadir) [copy] (number)
	9 [copy] (number)

number new datadir with basename copy from datadir

	9 (datadir) [copy] (basename) (number)
	9 [copy] (basename) (number)
		
1 new datadir with basename		

	9 (datadir) (basename)
	9 [new] (basename)

run 1 or more test nodes at ./test[0-9]{0,3} (all if no number) (to logfile)

	9 (test)
	9 (test) (number]
	9 (test) (number] (logfilepath)

run 1 or more test nodes at ./basename[0-9]{0,3} (all if no number) (to logfile)

	9 (test) (basename) 
	9 (test) (basename) (number]
	9 (test) (basename) (number] (logfilepath)

> # Note

- names are all [a-z][a-z0-9.-]+[a-z0-9]
 
- paths/dirs all must have / after them to easily distinguish

- all items in [] are keywords and must exactly match and be surrounded by spaces
 
- because none of the items are ambiguous there won't be false negatives from properly formed strings.