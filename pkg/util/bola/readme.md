# bola

## Reliable UDP Short Message Transport

>   #
>   #### **bola** [ boh-luh ]
>   
>  noun, plural bo·las [boh-luh z] /ˈboʊ ləz/.
>   
> 1. Also bolas. a weapon consisting of two or more heavy balls secured to the ends of one or more strong cords, hurled by the Indians and gauchos of southern South America to entangle the legs of cattle and other animals.
>   
>   https://www.dictionary.com/browse/bola?s=t
>   #

Many types of network services involve very short message chatter. Instant messaging, multiplayer games, increasingly websites are becoming asynchronous, push-updating with websockets and the Quic protocol by Google, using light RPC protocols or simply providing a small set of simple queries. These connectionless protocols simply fire off packets and prays they get to the other side in the right order and intact, and 

This library is super simple. It just provides the ability to send up to 3kb of data in one burst as 9 parts of which three valid pieces reconstructs the original message, and has a receiving processor that grabs a message as quickly as it can get 3 valid pieces it then starts attempting to assemble the message to then send to a message channel for consumption by the server.

There is no multiplexing, all messages are in a uniform format and should get through in the shortest possible time, in cases of congestion or packet loss, maybe at all.

The shards of the messages are ordered, but this is because they must occupy a particular position due to the fact the points, the pieces of the document, represent a coordinate on a polynomial curve. 

Bola does not concern itself with flow control for streams, handshakes, acks, or bulk transfers. These will be implemented with bola simply as the transport layer, bola's only purpose is to send and receive packets on a peer to peer basis between one application and another.