# Battle Bit
***

Who will be the first to set all the bits to 1?
Challenge your friends and AI bots to be the last one standing.

## How start ??

The server, running on port 8080 under the /ws endpoint, upgrades connections to a WebSocket implementation with JSON-RPC 2.0.

To create a game, connect to the server and send the following JSON request:
****
```JSON
{
    "jsonrpc": "2.0",
    "method": "create_game",
    "params": {
        "size": 1000000,
        "autopilots": 5
    },
    "id": "1"
}
```

This request will create a game with 1 million bits, and it will also include 5 AI bots that will compete against the human players.

Next, you need to join the game using the join_game method:
```JSON
{
    "jsonrpc": "2.0",
    "method": "join_game",
    "params": {
        "gameId": "uuid",
        "playerName": "JohnDoe"
    },
    "id": 1
}
```
After joining the game, you can strategize and make your moves with the following request:

```JSON
{
    "jsonrpc": "2.0",
    "method": "player_move",
    "params": {
        "gameId": "game-uuid",
        "playerId": "player-uuid",
        "index": 5
    },
    "id": 2
}
```

# Explore the Game and enjoy!!!