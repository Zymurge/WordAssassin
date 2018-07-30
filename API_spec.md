# WordAssassin Service

A web service that supports an online game of Word Assassin. The intent is to hook this up to Slack to automate things like authentication, secret passing, etc. In theory,
the game can even be extended into words typed into Slack, assuming there are APIs that can validate content.

## How To Play

A game is made up of a group of players each trying to assassinate the rest of the group. The winner is the last standing.

Once the setup phase is completed, each person is assigned a target (to asassinate) and a kill word. Both the target and the kill word are secret and known only to the assassin.

In order to assassinate the target, the assassin must get them to speak (or possibly type into Slack) that specific word. The assassin cannot say the word themselves in the context of the conversation. When an assassination is successfully completed, the target enters their death to the service. The service then passes on the current target of their just killed victim as their new assignment.

Once a player has been killed, the service status will publish who the assassin was and the kill word.

The game ends when the 2nd to last person registers their death, leaving the last assassin standing as the winner.

## APIs

- ###  **Status** *game-id*
        Game: <name>
        Status: {"starting" "running" "finished"}
        Time Running: dd hh:mm
  
        Players:
            Starting: num
            Alive: num

- ###  CreateGame *game-id creator kill-dictionary passcode*
  
- ###  AddPlayer *game-id player-tag*

- ###  StartGame *game-id*

- ###  ReportKill *game-id assassinated-by*

- ###  Target
