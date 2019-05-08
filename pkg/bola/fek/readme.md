# fek

## Forward Error Korrektion

*...because there is nothing more correct than korrekt! (**auf deutsch**)* *

`fek` is a super simple, performance focused library to generate arbitrary
Reed Solomon data and parity *frags* and from (externally verified) known good
frags it can reassemble the original message.

The data includes space for a 32 bit UUID value to distinguish between
groups of frags passing through a single channel, it is considered sufficient
as the intended use case for this library is a minimal latency UDP work dispatch
protocol. It can create frags up to the size of free, allocatable system memory
and therefore effectively can be used for any data that can fit in system memory
including large files.

Instead of employing error correction algorithms like
Peterson–Gorenstein–Zierler, Berlekamp–Massey or other, that attempts to
identify correct frags by their mutual consistency or other similar method, it
is left up to the caller to handle error detection and only to pass known valid
frags, at least the required by the preset Reed Solomon codec settings.

The purpose of this is to use fast checksums and enable minimal latency but
corruption/congestion resistant such as for a minimal latency network transport
or even a filesystem, since the cost of verification is proportional to the data
size for the checksum, but this is left up to the caller to determine.

`fek` rejects frags shorter than the required 32 bit UUID and 1 byte frag number
in any case, as well as bundles with a frag of differing length, and of less
than the required pieces.

This is a critical section codebase and maximum optimisation should be targeted,
especially to conserve on heap allocations, to further minimise its cost.

> ### footnote
> \* and no such thing as politically correct to the apolitical. See LICENCE for 
more info