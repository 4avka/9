# config

It is deceptively small but this library provides a neat and concise syntax for
declaration of a configuration data structure.

In the [cmd/](cmd/) folder is an example of a declaration of the struct and a
function that binds a struct containing pointers to the values into the
generated map of configuration values.

These are the various facilities that this design pattern provides:

1. Declaration is short and neat and readable and self explanatory
2. Produced data structure can be converted to json and decoded from a json
configuration file which includes useful information for a human editor in the
form of the constraints and usage text that apply from the declarattion
3. Produced structure contains initialisers, getters and setters that validate
all input.

If it was needed the structure can have mutex locks for concurrent read/write by
chaining unlock/lock into the initial validator and accessors, and chain onto
a channel notifying of changes in the configuration that can alter runtime
parameters, triggering a reinitialisation or so.

In its current form it makes specifying configuration just two functions that
mostly explain themselves, and provide a human readable structured document
matching the specification, to be written and read in a configuration file.