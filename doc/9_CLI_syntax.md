# <img src="https://git.parallelcoin.io/com/assets/raw/branch/master/logo/logo64x64.png"> <sup><i>9</i></sup> 
> **Parallelcoin** CLI and Configuration documentation

#### all the things it can do

## general description

In the design of the configuration and CLI parsing all conventions have been eschewed and the design is aimed at simplicity, maintainability, and relevant to the modern computer user.

Most flags never get used, and some people love to edit configuration files. If someone really wants 50 silly `--someridiculouslylongname` flags I suppose they can have that. As such, command line and configuration files are as simple as they can be while keeping a respectable amount of functionality. Such a feature is welcome to be added so long as it hides behind a subcommand.

## configuration file format

Regular expression looks like this

```
^(NAME)(.*)(VALUE)(.*)(DEFAULT)(.*)(TYPE)(.*)(COMMENT)(.*)$
```

Each item is separated by carriage returns, and these keywords act as visible markers cutting up the remainder of the string as 5 sections ending with the line ending. As such they can contain interstitial whitespace characters, anything, except of course those 5 words in that case.

The format is designed to be human readable and simple to parse. It should be obvious immediately what the format is when you look at it. Be careful with editors automatically adding carriage returns. The comment fields are not that long and come last, nothing follows.

If the parser hits an error it will show a warning but the failed lines will have default values.

### language processors are expensive in man hours

Rather than waste time making clever contingencies with fancy state machines for no reason, a certain competence is expected from the user but nobody has time to learn the book of `man`, verse `3`, psalm `69`...

Instead, you can just explore the search as it narrows down matches as you add characters, select an item, and its full help information will appear, and you can immediately change it. It is exceedingly inefficient for a programmer to have to attend to any task more than once or with more than 10 steps in a row without breaking.

### self documenting

The configuration file contains the brief but not uninformative short usage texts, so the document is somewhat informative by itself.

The declarations are compact and quite clearly denote the meaning of the parameters.

Command line flags do not profit thee more with their plurality. If you are filling more than 80 columns it probably won't take any longer to use a type-ahead search and tree navigator, or a simple keyword separated file in an editor.

Command line parsing works by a principle of identifiable patterns matching a specified set. There is no real need for anything more complex than that, configuration interfaces are better than complex giant long command lines. Thus no command line contains more than one type of thing except it can end in an arbitrary string for subcommands.

## command line flags (exhaustive)

launch gui (default datadir)

    9

launch gui (other datadir)

    9 {datadir}

show help

    9 {datadir} [help|h]
    9 [help|h]

launch configuration cli

    9 {datadir} [conf|c]
    9 [conf|c]

show listcommands

    9 {datadir} (listcommands|l)
    9 [listcommands|l]

query from full node

	9 {datadir} [rpc|r] (getinfo/blahblah...)
	9 [rpc|r] (getinfo/blahblah...)

run full node

	9 {datadir} [node|n]
	9 [node|n]

run wallet

	9 {datadir} [node|n]
	9 [node|n]

run shell

	9 {datadir} [shell|s]
	9 [shell|s]

reset to factory defaults

	9 {datadir} [factory]
	9 [factory]

new datadir at ./test

	9 [new]

new datadir default at ./test

	9 [new] (number)

number new datadir with basename

	9 [new] (basename) (number)
		
1 new datadir with basename		

	9 [new] (basename)

copy from home to datadir at ./test

	9 {datadir} [copy]
	9 [copy]

copy datadir at ./test

	9 {datadir} [copy] (number)
	9 [copy] (number)

number new datadir with basename copy from datadir

	9 {datadir} [copy] (basename) (number)
	9 [copy] (basename) (number)
		
1 new datadir with basename		

	9 {datadir} (basename)
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

- datadir is ALWAYS first parameter, all the others can appear in any other position.

## final notes

CLI libraries are supposed to save us time. But in the time it takes to figure the stupid things out you can probably find if you don't slavishly adhere to the flag for every widget philosophy. If they fit, they will eventually find their way into it.